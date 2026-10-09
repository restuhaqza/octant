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

func grpcRouteWithRefs() *gatewayv1.GRPCRoute {
	route := testutil.CreateGRPCRoute("grpc-route")
	route.Spec.Hostnames = []gatewayv1.Hostname{"example.com"}
	route.Spec.ParentRefs = []gatewayv1.ParentReference{
		{Name: "gateway"},
	}
	route.Spec.Rules = []gatewayv1.GRPCRouteRule{
		{
			BackendRefs: []gatewayv1.GRPCBackendRef{
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
	return route
}

func Test_GRPCRouteListHandler(t *testing.T) {
	cols := component.NewTableCols("Name", "Hostnames", "Parents", "Backends", "Age")
	now := testutil.Time()

	object := grpcRouteWithRefs()
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &gatewayv1.GRPCRouteList{
		Items: []gatewayv1.GRPCRoute{*object},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	tpo := newTestPrinterOptions(controller)
	printOptions := tpo.ToOptions()

	tpo.PathForObject(object, object.Name, "/grpc-route")
	ctx := context.Background()
	tpo.pluginManager.EXPECT().ObjectStatus(ctx, object)

	got, err := GRPCRouteListHandler(ctx, list, printOptions)
	require.NoError(t, err)

	expected := component.NewTableWithRows("GRPC Routes", "We couldn't find any grpc routes!", cols,
		[]component.TableRow{
			{
				"Name": component.NewLink("", "grpc-route", "/grpc-route",
					genObjectStatus(component.TextStatusWarning, []string{
						"Route is not attached to any parent",
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

func Test_GRPCRouteConfiguration(t *testing.T) {
	route := grpcRouteWithRefs()

	controller := gomock.NewController(t)
	defer controller.Finish()

	tpo := newTestPrinterOptions(controller)
	printOptions := tpo.ToOptions()

	got, err := NewGRPCRouteConfiguration(route).Create(printOptions)
	require.NoError(t, err)

	expected := component.NewSummary("Configuration", []component.SummarySection{
		{Header: "Hostnames", Content: component.NewText("example.com")},
		{Header: "Parents", Content: component.NewText("Gateway namespace/gateway")},
		{Header: "Accepted", Content: component.NewText("Unknown")},
		{Header: "Resolved Refs", Content: component.NewText("Unknown")},
	}...)

	component.AssertEqual(t, expected, got)
}
