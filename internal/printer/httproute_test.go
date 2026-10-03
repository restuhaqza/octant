/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/vmware-tanzu/octant/internal/testutil"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

func httpRouteWithRefs() *gatewayv1.HTTPRoute {
	route := testutil.CreateHTTPRoute("http-route")
	route.Spec.Hostnames = []gatewayv1.Hostname{"example.com"}
	route.Spec.ParentRefs = []gatewayv1.ParentReference{
		{Name: "gateway"},
	}
	route.Spec.Rules = []gatewayv1.HTTPRouteRule{
		{
			BackendRefs: []gatewayv1.HTTPBackendRef{
				{
					BackendRef: gatewayv1.BackendRef{
						BackendObjectReference: gatewayv1.BackendObjectReference{
							Name: "service",
						},
					},
				},
			},
		},
	}
	route.Status.Parents = []gatewayv1.RouteParentStatus{
		{
			ParentRef: gatewayv1.ParentReference{Name: "gateway"},
			Conditions: []metav1.Condition{
				{Type: string(gatewayv1.RouteConditionAccepted), Status: metav1.ConditionTrue},
				{Type: string(gatewayv1.RouteConditionResolvedRefs), Status: metav1.ConditionTrue},
			},
		},
	}
	return route
}

func Test_HTTPRouteListHandler(t *testing.T) {
	cols := component.NewTableCols("Name", "Hostnames", "Parents", "Backends", "Age")
	now := testutil.Time()

	object := httpRouteWithRefs()
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &gatewayv1.HTTPRouteList{
		Items: []gatewayv1.HTTPRoute{*object},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	tpo := newTestPrinterOptions(controller)
	printOptions := tpo.ToOptions()

	tpo.PathForObject(object, object.Name, "/http-route")
	ctx := context.Background()
	tpo.pluginManager.EXPECT().ObjectStatus(ctx, object)

	got, err := HTTPRouteListHandler(ctx, list, printOptions)
	require.NoError(t, err)

	expected := component.NewTableWithRows("HTTP Routes", "We couldn't find any http routes!", cols,
		[]component.TableRow{
			{
				"Name": component.NewLink("", "http-route", "/http-route",
					genObjectStatus(component.TextStatusOK, []string{
						`Accepted by parent "gateway"`,
					})),
				"Hostnames": component.NewText("example.com"),
				"Parents":   component.NewText("Gateway namespace/gateway"),
				"Backends":  component.NewText("Service namespace/service"),
				"Age":       component.NewTimestamp(now),
				component.GridActionKey: gridActionsFactory([]component.GridAction{
					buildObjectDeleteAction(t, object),
				}),
			},
		})

	component.AssertEqual(t, expected, got)
}

func Test_HTTPRouteConfiguration(t *testing.T) {
	route := httpRouteWithRefs()

	controller := gomock.NewController(t)
	defer controller.Finish()

	tpo := newTestPrinterOptions(controller)
	printOptions := tpo.ToOptions()

	got, err := NewHTTPRouteConfiguration(route).Create(printOptions)
	require.NoError(t, err)

	expected := component.NewSummary("Configuration", []component.SummarySection{
		{Header: "Hostnames", Content: component.NewText("example.com")},
		{Header: "Parents", Content: component.NewText("Gateway namespace/gateway")},
		{Header: "Accepted", Content: component.NewText("True")},
		{Header: "Resolved Refs", Content: component.NewText("True")},
	}...)

	component.AssertEqual(t, expected, got)
}
