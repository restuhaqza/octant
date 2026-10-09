/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectstatus

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/vmware-tanzu/octant/internal/link"
	"github.com/vmware-tanzu/octant/pkg/store"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

const nodeRoleLabelPrefix = "node-role.kubernetes.io/"

// node creates status for a v1 Node. Without this the node status falls back to
// the generic "... is OK" and a NotReady or cordoned node looks healthy.
func node(_ context.Context, object runtime.Object, _ store.Store, _ link.Interface) (ObjectStatus, error) {
	if object == nil {
		return ObjectStatus{}, errors.Errorf("node is nil")
	}

	node := &corev1.Node{}
	if err := scheme.Scheme.Convert(object, node, 0); err != nil {
		return ObjectStatus{}, errors.Wrap(err, "convert object to node")
	}

	var status ObjectStatus

	switch ready := findNodeCondition(node.Status.Conditions, corev1.NodeReady); {
	case ready == nil:
		status.SetWarning()
		status.AddDetail("Node has not reported a Ready condition")
	case ready.Status == corev1.ConditionTrue:
		status.AddDetail("Node is ready")
	case ready.Status == corev1.ConditionUnknown:
		status.SetWarning()
		status.AddDetailf("Node readiness is unknown: (%s) %s", ready.Reason, ready.Message)
	default:
		status.SetError()
		status.AddDetailf("Node is not ready: (%s) %s", ready.Reason, ready.Message)
	}

	for _, pressure := range []struct {
		condition corev1.NodeConditionType
		label     string
	}{
		{corev1.NodeMemoryPressure, "memory pressure"},
		{corev1.NodeDiskPressure, "disk pressure"},
		{corev1.NodePIDPressure, "PID pressure"},
		{corev1.NodeNetworkUnavailable, "network unavailable"},
	} {
		if c := findNodeCondition(node.Status.Conditions, pressure.condition); c != nil && c.Status == corev1.ConditionTrue {
			status.SetWarning()
			status.AddDetailf("Node reports %s", pressure.label)
		}
	}

	if node.Spec.Unschedulable {
		status.SetWarning()
		status.AddDetail("Scheduling is disabled (node is cordoned)")
	}

	status.AddProperty("Roles", component.NewText(nodeRoles(node)))
	status.AddProperty("Kubelet Version", component.NewText(node.Status.NodeInfo.KubeletVersion))
	status.AddProperty("OS/Arch", component.NewText(fmt.Sprintf("%s/%s", node.Status.NodeInfo.OperatingSystem, node.Status.NodeInfo.Architecture)))
	status.AddProperty("Internal IP", component.NewText(nodeAddress(node, corev1.NodeInternalIP)))
	status.AddProperty("CPU", component.NewText(node.Status.Capacity.Cpu().String()))
	status.AddProperty("Memory", component.NewText(node.Status.Capacity.Memory().String()))
	status.AddProperty("Pod CIDR", component.NewText(node.Spec.PodCIDR))

	return status, nil
}

func findNodeCondition(conditions []corev1.NodeCondition, conditionType corev1.NodeConditionType) *corev1.NodeCondition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}

func nodeRoles(node *corev1.Node) string {
	seen := make(map[string]bool)
	var roles []string

	for label := range node.Labels {
		if !strings.HasPrefix(label, nodeRoleLabelPrefix) {
			continue
		}
		role := strings.TrimPrefix(label, nodeRoleLabelPrefix)
		if role == "" || seen[role] {
			continue
		}
		seen[role] = true
		roles = append(roles, role)
	}

	if len(roles) == 0 {
		return "<none>"
	}

	sort.Strings(roles)
	return strings.Join(roles, ", ")
}

func nodeAddress(node *corev1.Node, addressType corev1.NodeAddressType) string {
	for _, address := range node.Status.Addresses {
		if address.Type == addressType {
			return address.Address
		}
	}
	return "<none>"
}
