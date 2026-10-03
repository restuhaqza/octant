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
	"k8s.io/client-go/kubernetes/scheme"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/vmware-tanzu/octant/internal/testutil"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

func gatewayWithConds() *gatewayv1.Gateway {
	gateway := testutil.CreateGateway("gateway")
	gateway.Spec.Listeners = []gatewayv1.Listener{
		{
			Name:     "http",
			Port:     80,
			Protocol: gatewayv1.HTTPProtocolType,
		},
	}
	gateway.Status.Conditions = []metav1.Condition{
		{Type: string(gatewayv1.GatewayConditionAccepted), Status: metav1.ConditionTrue},
		{Type: string(gatewayv1.GatewayConditionProgrammed), Status: metav1.ConditionTrue},
	}
	return gateway
}

func Test_GatewayListHandler(t *testing.T) {
	cols := component.NewTableCols("Name", "Class", "Listeners", "Addresses", "Age")
	now := testutil.Time()

	object := gatewayWithConds()
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &gatewayv1.GatewayList{
		Items: []gatewayv1.Gateway{*object},
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	tpo := newTestPrinterOptions(controller)
	printOptions := tpo.ToOptions()

	tpo.PathForObject(object, object.Name, "/gateway")
	ctx := context.Background()
	tpo.pluginManager.EXPECT().ObjectStatus(ctx, object)

	got, err := GatewayListHandler(ctx, list, printOptions)
	require.NoError(t, err)

	expected := component.NewTableWithRows("Gateways", "We couldn't find any gateways!", cols,
		[]component.TableRow{
			{
				"Name": component.NewLink("", "gateway", "/gateway",
					genObjectStatus(component.TextStatusOK, []string{
						"Gateway is OK",
					})),
				"Class":     component.NewText("gateway-class"),
				"Listeners": component.NewText("1"),
				"Addresses": component.NewText("<none>"),
				"Age":       component.NewTimestamp(now),
				component.GridActionKey: gridActionsFactory([]component.GridAction{
					buildObjectDeleteAction(t, object),
				}),
			},
		})

	component.AssertEqual(t, expected, got)
}

func Test_GatewayConfiguration(t *testing.T) {
	gateway := gatewayWithConds()

	controller := gomock.NewController(t)
	defer controller.Finish()

	tpo := newTestPrinterOptions(controller)
	printOptions := tpo.ToOptions()

	tpo.PathForGVK("", "gateway.networking.k8s.io/v1", "GatewayClass", "gateway-class", "gateway-class", "/gateway-class")

	got, err := NewGatewayConfiguration(gateway).Create(printOptions)
	require.NoError(t, err)

	expected := component.NewSummary("Configuration", []component.SummarySection{
		{Header: "Gateway Class", Content: component.NewLink("", "gateway-class", "/gateway-class")},
		{Header: "Listeners", Content: component.NewText("1")},
		{Header: "Addresses", Content: component.NewText("<none>")},
		{Header: "Accepted", Content: component.NewText("True")},
		{Header: "Programmed", Content: component.NewText("True")},
	}...)

	component.AssertEqual(t, expected, got)
}

func Test_createGatewayListenersView(t *testing.T) {
	gateway := gatewayWithConds()

	got, err := createGatewayListenersView(gateway)
	require.NoError(t, err)

	cols := component.NewTableCols("Name", "Hostname", "Port", "Protocol")
	expected := component.NewTable("Listeners", "There are no listeners defined!", cols)
	expected.Add(component.TableRow{
		"Name":     component.NewText("http"),
		"Hostname": component.NewText("*"),
		"Port":     component.NewText("80"),
		"Protocol": component.NewText("HTTP"),
	})

	component.AssertEqual(t, expected, got)
}

func init() {
	_ = gatewayv1.Install(scheme.Scheme)
}
