/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectstatus

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/vmware-tanzu/octant/internal/link"
	"github.com/vmware-tanzu/octant/pkg/store"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// gateway creates status for a gateway.networking.k8s.io/v1 Gateway.
func gateway(_ context.Context, object runtime.Object, _ store.Store, _ link.Interface) (ObjectStatus, error) {
	if object == nil {
		return ObjectStatus{}, errors.Errorf("gateway is nil")
	}

	gateway := &gatewayv1.Gateway{}
	if err := scheme.Scheme.Convert(object, gateway, 0); err != nil {
		return ObjectStatus{}, errors.Wrap(err, "convert object to gateway")
	}

	var status ObjectStatus

	switch condition := findCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)); {
	case condition == nil:
		status.SetWarning()
		status.AddDetail("No accepted condition for this gateway")
	case condition.Status == metav1.ConditionFalse:
		status.SetError()
		status.AddDetailf("Not accepted: (%s) %s", condition.Reason, condition.Message)
	case condition.Status == metav1.ConditionUnknown:
		status.SetWarning()
		status.AddDetailf("Acceptance unknown: (%s) %s", condition.Reason, condition.Message)
	default:
		status.AddDetail("Gateway is OK")
	}

	// Surface a programmed failure if the gateway has not been programmed.
	if condition := findCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionProgrammed)); condition != nil && condition.Status == metav1.ConditionFalse {
		status.SetError()
		status.AddDetailf("Not programmed: (%s) %s", condition.Reason, condition.Message)
	}

	addresses := "<none>"
	if len(gateway.Status.Addresses) > 0 {
		var list []string
		for _, address := range gateway.Status.Addresses {
			list = append(list, address.Value)
		}
		addresses = strings.Join(list, ", ")
	}

	status.AddProperty("Gateway Class", component.NewText(string(gateway.Spec.GatewayClassName)))
	status.AddProperty("Listeners", component.NewText(fmt.Sprintf("%d", len(gateway.Spec.Listeners))))
	status.AddProperty("Addresses", component.NewText(addresses))

	return status, nil
}

// httpRoute creates status for a gateway.networking.k8s.io/v1 HTTPRoute.
func httpRoute(_ context.Context, object runtime.Object, _ store.Store, _ link.Interface) (ObjectStatus, error) {
	if object == nil {
		return ObjectStatus{}, errors.Errorf("http route is nil")
	}

	route := &gatewayv1.HTTPRoute{}
	if err := scheme.Scheme.Convert(object, route, 0); err != nil {
		return ObjectStatus{}, errors.Wrap(err, "convert object to http route")
	}

	var status ObjectStatus

	if len(route.Status.Parents) == 0 {
		status.SetWarning()
		status.AddDetail("Route is not attached to any parent")
	} else {
		for _, parent := range route.Status.Parents {
			name := string(parent.ParentRef.Name)

			switch condition := findCondition(parent.Conditions, string(gatewayv1.RouteConditionAccepted)); {
			case condition == nil:
				status.SetWarning()
				status.AddDetailf("No accepted condition for parent %q", name)
			case condition.Status == metav1.ConditionFalse:
				status.SetError()
				status.AddDetailf("Not accepted by parent %q: (%s) %s", name, condition.Reason, condition.Message)
			case condition.Status == metav1.ConditionUnknown:
				status.SetWarning()
				status.AddDetailf("Acceptance unknown for parent %q: (%s) %s", name, condition.Reason, condition.Message)
			default:
				status.AddDetailf("Accepted by parent %q", name)
			}

			if condition := findCondition(parent.Conditions, string(gatewayv1.RouteConditionResolvedRefs)); condition != nil && condition.Status == metav1.ConditionFalse {
				status.SetError()
				status.AddDetailf("Unresolved refs for parent %q: (%s) %s", name, condition.Reason, condition.Message)
			}
		}
	}

	if len(status.Details) == 0 {
		status.AddDetail("HTTPRoute is OK")
	}

	status.AddProperty("Hostnames", component.NewText(formatRouteHostnames(route.Spec.Hostnames)))
	status.AddProperty("Parents", component.NewText(fmt.Sprintf("%d", len(route.Spec.ParentRefs))))
	status.AddProperty("Rules", component.NewText(fmt.Sprintf("%d", len(route.Spec.Rules))))

	return status, nil
}

func findCondition(conditions []metav1.Condition, conditionType string) *metav1.Condition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}

func formatRouteHostnames(hostnames []gatewayv1.Hostname) string {
	if len(hostnames) == 0 {
		return "*"
	}

	list := make([]string, 0, len(hostnames))
	for _, hostname := range hostnames {
		list = append(list, string(hostname))
	}
	return strings.Join(list, ", ")
}
