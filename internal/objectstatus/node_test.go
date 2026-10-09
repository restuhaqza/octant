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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/octant/internal/testutil"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

func Test_node(t *testing.T) {
	readyNode := func() *corev1.Node {
		node := testutil.CreateNode("node")
		node.Status.Conditions = []corev1.NodeCondition{
			{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
		}
		node.Status.NodeInfo = corev1.NodeSystemInfo{
			KubeletVersion:  "v1.34.1",
			OperatingSystem: "linux",
			Architecture:    "amd64",
		}
		node.Status.Addresses = []corev1.NodeAddress{
			{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
		}
		node.Labels = map[string]string{"node-role.kubernetes.io/control-plane": ""}
		node.Status.Capacity = corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("4"),
			corev1.ResourceMemory: resource.MustParse("8Gi"),
		}
		node.Spec.PodCIDR = "10.244.0.0/24"
		return node
	}

	properties := []component.Property{
		{Label: "Roles", Value: component.NewText("control-plane")},
		{Label: "Kubelet Version", Value: component.NewText("v1.34.1")},
		{Label: "OS/Arch", Value: component.NewText("linux/amd64")},
		{Label: "Internal IP", Value: component.NewText("10.0.0.1")},
		{Label: "CPU", Value: component.NewText("4")},
		{Label: "Memory", Value: component.NewText("8Gi")},
		{Label: "Pod CIDR", Value: component.NewText("10.244.0.0/24")},
	}

	emptyProperties := []component.Property{
		{Label: "Roles", Value: component.NewText("<none>")},
		{Label: "Kubelet Version", Value: component.NewText("")},
		{Label: "OS/Arch", Value: component.NewText("/")},
		{Label: "Internal IP", Value: component.NewText("<none>")},
		{Label: "CPU", Value: component.NewText("0")},
		{Label: "Memory", Value: component.NewText("0")},
		{Label: "Pod CIDR", Value: component.NewText("")},
	}

	cases := []struct {
		name     string
		object   runtime.Object
		expected ObjectStatus
		isErr    bool
	}{
		{
			name:   "ready",
			object: readyNode(),
			expected: ObjectStatus{
				Details:    []component.Component{component.NewText("Node is ready")},
				Properties: properties,
			},
		},
		{
			name: "not ready",
			object: func() runtime.Object {
				node := readyNode()
				node.Status.Conditions = []corev1.NodeCondition{
					{Type: corev1.NodeReady, Status: corev1.ConditionFalse, Reason: "KubeletNotReady", Message: "runtime down"},
				}
				return node
			}(),
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusError,
				Details:    []component.Component{component.NewText("Node is not ready: (KubeletNotReady) runtime down")},
				Properties: properties,
			},
		},
		{
			name: "memory pressure and cordoned",
			object: func() runtime.Object {
				node := readyNode()
				node.Status.Conditions = append(node.Status.Conditions,
					corev1.NodeCondition{Type: corev1.NodeMemoryPressure, Status: corev1.ConditionTrue})
				node.Spec.Unschedulable = true
				return node
			}(),
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusWarning,
				Details: []component.Component{
					component.NewText("Node is ready"),
					component.NewText("Node reports memory pressure"),
					component.NewText("Scheduling is disabled (node is cordoned)"),
				},
				Properties: properties,
			},
		},
		{
			name:   "no conditions",
			object: testutil.CreateNode("node"),
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusWarning,
				Details:    []component.Component{component.NewText("Node has not reported a Ready condition")},
				Properties: emptyProperties,
			},
		},
		{
			name:   "nil object",
			object: nil,
			isErr:  true,
		},
		{
			name:   "not a node",
			object: &unstructured.Unstructured{},
			isErr:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := node(context.Background(), tc.object, nil, nil)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}
