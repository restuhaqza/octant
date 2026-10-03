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

func Test_GatewayClassListHandler(t *testing.T) {
	cols := component.NewTableCols("Name", "Controller", "Accepted", "Age")
	now := testutil.Time()

	object := testutil.CreateGatewayClass("gateway-class")
	object.CreationTimestamp = metav1.Time{Time: now}
	object.Status.Conditions = []metav1.Condition{
		{
			Type:   string(gatewayv1.GatewayClassConditionStatusAccepted),
			Status: metav1.ConditionTrue,
		},
	}

	list := &gatewayv1.GatewayClassList{
		Items: []gatewayv1.GatewayClass{*object},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	tpo := newTestPrinterOptions(controller)
	printOptions := tpo.ToOptions()

	tpo.PathForObject(object, object.Name, "/gateway-class")
	ctx := context.Background()
	tpo.pluginManager.EXPECT().ObjectStatus(ctx, object)

	got, err := GatewayClassListHandler(ctx, list, printOptions)
	require.NoError(t, err)

	expected := component.NewTableWithRows("Gateway Classes", "We couldn't find any gateway classes!", cols,
		[]component.TableRow{
			{
				"Name": component.NewLink("", "gateway-class", "/gateway-class",
					genObjectStatus(component.TextStatusOK, []string{
						"gateway.networking.k8s.io/v1 GatewayClass is OK",
					})),
				"Controller": component.NewText("example.com/controller"),
				"Accepted":   component.NewText("True"),
				"Age":        component.NewTimestamp(now),
				component.GridActionKey: gridActionsFactory([]component.GridAction{
					buildObjectDeleteAction(t, object),
				}),
			},
		})

	component.AssertEqual(t, expected, got)
}

func Test_GatewayClassConfiguration(t *testing.T) {
	gatewayClass := testutil.CreateGatewayClass("gateway-class")
	description := "a gateway class"
	gatewayClass.Spec.Description = &description

	controller := gomock.NewController(t)
	defer controller.Finish()

	tpo := newTestPrinterOptions(controller)
	printOptions := tpo.ToOptions()

	got, err := NewGatewayClassConfiguration(gatewayClass).Create(printOptions)
	require.NoError(t, err)

	expected := component.NewSummary("Configuration", []component.SummarySection{
		{Header: "Controller", Content: component.NewText("example.com/controller")},
		{Header: "Accepted", Content: component.NewText("Unknown")},
		{Header: "Description", Content: component.NewText(description)},
	}...)

	component.AssertEqual(t, expected, got)
}
