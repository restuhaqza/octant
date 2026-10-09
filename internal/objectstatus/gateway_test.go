/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectstatus

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/vmware-tanzu/octant/internal/testutil"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

func Test_gateway(t *testing.T) {
	withConditions := func(conditions ...metav1.Condition) *gatewayv1.Gateway {
		gateway := testutil.CreateGateway("gateway")
		gateway.Status.Conditions = conditions
		return gateway
	}

	cases := []struct {
		name     string
		object   runtime.Object
		expected ObjectStatus
		isErr    bool
	}{
		{
			name: "accepted and programmed",
			object: withConditions(
				metav1.Condition{Type: string(gatewayv1.GatewayConditionAccepted), Status: metav1.ConditionTrue},
				metav1.Condition{Type: string(gatewayv1.GatewayConditionProgrammed), Status: metav1.ConditionTrue},
			),
			expected: ObjectStatus{
				Details: []component.Component{component.NewText("Gateway is OK")},
				Properties: []component.Property{
					{Label: "Gateway Class", Value: component.NewText("gateway-class")},
					{Label: "Listeners", Value: component.NewText("0")},
					{Label: "Addresses", Value: component.NewText("<none>")},
				},
			},
		},
		{
			name: "not accepted",
			object: withConditions(
				metav1.Condition{
					Type:    string(gatewayv1.GatewayConditionAccepted),
					Status:  metav1.ConditionFalse,
					Reason:  "Pending",
					Message: "not ready",
				},
			),
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusError,
				Details:    []component.Component{component.NewText("Not accepted: (Pending) not ready")},
				Properties: []component.Property{
					{Label: "Gateway Class", Value: component.NewText("gateway-class")},
					{Label: "Listeners", Value: component.NewText("0")},
					{Label: "Addresses", Value: component.NewText("<none>")},
				},
			},
		},
		{
			name:   "no conditions",
			object: withConditions(),
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusWarning,
				Details:    []component.Component{component.NewText("No accepted condition for this gateway")},
				Properties: []component.Property{
					{Label: "Gateway Class", Value: component.NewText("gateway-class")},
					{Label: "Listeners", Value: component.NewText("0")},
					{Label: "Addresses", Value: component.NewText("<none>")},
				},
			},
		},
		{
			name:   "nil object",
			object: nil,
			isErr:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			got, err := gateway(ctx, tc.object, nil, nil)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func Test_httpRoute(t *testing.T) {
	route := testutil.CreateHTTPRoute("http-route")
	route.Spec.Hostnames = []gatewayv1.Hostname{"example.com"}
	route.Spec.ParentRefs = []gatewayv1.ParentReference{
		{Name: "gateway"},
	}
	route.Status.Parents = []gatewayv1.RouteParentStatus{
		{
			ParentRef: gatewayv1.ParentReference{Name: "gateway"},
			Conditions: []metav1.Condition{
				{Type: string(gatewayv1.RouteConditionAccepted), Status: metav1.ConditionTrue},
			},
		},
	}

	unattached := testutil.CreateHTTPRoute("http-route")
	unattached.Spec.Hostnames = []gatewayv1.Hostname{"example.com"}

	rejected := testutil.CreateHTTPRoute("http-route")
	rejected.Status.Parents = []gatewayv1.RouteParentStatus{
		{
			ParentRef: gatewayv1.ParentReference{Name: "gateway"},
			Conditions: []metav1.Condition{
				{
					Type:    string(gatewayv1.RouteConditionAccepted),
					Status:  metav1.ConditionFalse,
					Reason:  "NoMatchingParent",
					Message: "not allowed",
				},
			},
		},
	}

	baseProperties := []component.Property{
		{Label: "Hostnames", Value: component.NewText("example.com")},
	}

	cases := []struct {
		name     string
		object   runtime.Object
		expected ObjectStatus
		isErr    bool
	}{
		{
			name:   "accepted",
			object: route,
			expected: ObjectStatus{
				Details: []component.Component{component.NewText(`Accepted by parent "gateway"`)},
				Properties: append(baseProperties,
					component.Property{Label: "Parents", Value: component.NewText("1")},
					component.Property{Label: "Rules", Value: component.NewText("0")},
				),
			},
		},
		{
			name:   "unattached",
			object: unattached,
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusWarning,
				Details:    []component.Component{component.NewText("Route is not attached to any parent")},
				Properties: append(baseProperties,
					component.Property{Label: "Parents", Value: component.NewText("0")},
					component.Property{Label: "Rules", Value: component.NewText("0")},
				),
			},
		},
		{
			name:   "rejected",
			object: rejected,
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusError,
				Details:    []component.Component{component.NewText(`Not accepted by parent "gateway": (NoMatchingParent) not allowed`)},
				Properties: []component.Property{
					{Label: "Hostnames", Value: component.NewText("*")},
					{Label: "Parents", Value: component.NewText("0")},
					{Label: "Rules", Value: component.NewText("0")},
				},
			},
		},
		{
			name:   "nil object",
			object: nil,
			isErr:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			got, err := httpRoute(ctx, tc.object, nil, nil)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func Test_gatewayClass(t *testing.T) {
	accepted := testutil.CreateGatewayClass("gateway-class")
	accepted.Status.Conditions = []metav1.Condition{
		{Type: string(gatewayv1.GatewayClassConditionStatusAccepted), Status: metav1.ConditionTrue},
	}

	rejected := testutil.CreateGatewayClass("gateway-class")
	rejected.Status.Conditions = []metav1.Condition{
		{
			Type:    string(gatewayv1.GatewayClassConditionStatusAccepted),
			Status:  metav1.ConditionFalse,
			Reason:  "InvalidParameters",
			Message: "bad params",
		},
	}

	properties := []component.Property{
		{Label: "Controller", Value: component.NewText("example.com/controller")},
		{Label: "Description", Value: component.NewText("<none>")},
	}

	cases := []struct {
		name     string
		object   runtime.Object
		expected ObjectStatus
		isErr    bool
	}{
		{
			name:   "accepted",
			object: accepted,
			expected: ObjectStatus{
				Details:    []component.Component{component.NewText("GatewayClass is OK")},
				Properties: properties,
			},
		},
		{
			name:   "rejected",
			object: rejected,
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusError,
				Details:    []component.Component{component.NewText("Not accepted: (InvalidParameters) bad params")},
				Properties: properties,
			},
		},
		{
			name:   "no conditions",
			object: testutil.CreateGatewayClass("gateway-class"),
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusWarning,
				Details:    []component.Component{component.NewText("No accepted condition for this gateway class")},
				Properties: properties,
			},
		},
		{
			name:   "nil object",
			object: nil,
			isErr:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := gatewayClass(context.Background(), tc.object, nil, nil)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func Test_grpcRoute(t *testing.T) {
	route := testutil.CreateGRPCRoute("grpc-route")
	route.Spec.Hostnames = []gatewayv1.Hostname{"example.com"}
	route.Spec.ParentRefs = []gatewayv1.ParentReference{
		{Name: "gateway"},
	}
	route.Status.Parents = []gatewayv1.RouteParentStatus{
		{
			ParentRef: gatewayv1.ParentReference{Name: "gateway"},
			Conditions: []metav1.Condition{
				{Type: string(gatewayv1.RouteConditionAccepted), Status: metav1.ConditionTrue},
			},
		},
	}

	unattached := testutil.CreateGRPCRoute("grpc-route")
	unattached.Spec.Hostnames = []gatewayv1.Hostname{"example.com"}

	rejected := testutil.CreateGRPCRoute("grpc-route")
	rejected.Status.Parents = []gatewayv1.RouteParentStatus{
		{
			ParentRef: gatewayv1.ParentReference{Name: "gateway"},
			Conditions: []metav1.Condition{
				{
					Type:    string(gatewayv1.RouteConditionAccepted),
					Status:  metav1.ConditionFalse,
					Reason:  "NoMatchingParent",
					Message: "not allowed",
				},
			},
		},
	}

	baseProperties := []component.Property{
		{Label: "Hostnames", Value: component.NewText("example.com")},
	}

	cases := []struct {
		name     string
		object   runtime.Object
		expected ObjectStatus
		isErr    bool
	}{
		{
			name:   "accepted",
			object: route,
			expected: ObjectStatus{
				Details: []component.Component{component.NewText(`Accepted by parent "gateway"`)},
				Properties: append(baseProperties,
					component.Property{Label: "Parents", Value: component.NewText("1")},
					component.Property{Label: "Rules", Value: component.NewText("0")},
				),
			},
		},
		{
			name:   "unattached",
			object: unattached,
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusWarning,
				Details:    []component.Component{component.NewText("Route is not attached to any parent")},
				Properties: append(baseProperties,
					component.Property{Label: "Parents", Value: component.NewText("0")},
					component.Property{Label: "Rules", Value: component.NewText("0")},
				),
			},
		},
		{
			name:   "rejected",
			object: rejected,
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusError,
				Details:    []component.Component{component.NewText(`Not accepted by parent "gateway": (NoMatchingParent) not allowed`)},
				Properties: []component.Property{
					{Label: "Hostnames", Value: component.NewText("*")},
					{Label: "Parents", Value: component.NewText("0")},
					{Label: "Rules", Value: component.NewText("0")},
				},
			},
		},
		{
			name:   "nil object",
			object: nil,
			isErr:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := grpcRoute(context.Background(), tc.object, nil, nil)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func init() {
	_ = gatewayv1.Install(scheme.Scheme)
}
