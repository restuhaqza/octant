/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectstatus

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

func Test_horizontalPodAutoscaler(t *testing.T) {
	newHPA := func() *autoscalingv2.HorizontalPodAutoscaler {
		return &autoscalingv2.HorizontalPodAutoscaler{
			TypeMeta:   metav1.TypeMeta{APIVersion: "autoscaling/v2", Kind: "HorizontalPodAutoscaler"},
			ObjectMeta: metav1.ObjectMeta{Name: "hpa", Namespace: "namespace"},
			Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
				ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{Kind: "Deployment", Name: "app"},
				MaxReplicas:    5,
			},
		}
	}

	properties := func(current, desired int) []component.Property {
		return []component.Property{
			{Label: "Target", Value: component.NewText("Deployment/app")},
			{Label: "Minimum Replicas", Value: component.NewText("<none>")},
			{Label: "Maximum Replicas", Value: component.NewText("5")},
			{Label: "Current Replicas", Value: component.NewText(fmt.Sprintf("%d", current))},
			{Label: "Desired Replicas", Value: component.NewText(fmt.Sprintf("%d", desired))},
		}
	}

	cases := []struct {
		name     string
		object   runtime.Object
		expected ObjectStatus
		isErr    bool
	}{
		{
			name: "stable",
			object: func() runtime.Object {
				hpa := newHPA()
				hpa.Status.CurrentReplicas = 2
				hpa.Status.DesiredReplicas = 2
				return hpa
			}(),
			expected: ObjectStatus{
				Details:    []component.Component{component.NewText("Horizontal Pod Autoscaler is OK")},
				Properties: properties(2, 2),
			},
		},
		{
			name: "scaling not active",
			object: func() runtime.Object {
				hpa := newHPA()
				hpa.Status.CurrentReplicas = 2
				hpa.Status.DesiredReplicas = 2
				hpa.Status.Conditions = []autoscalingv2.HorizontalPodAutoscalerCondition{
					{
						Type:    autoscalingv2.ScalingActive,
						Status:  corev1.ConditionFalse,
						Reason:  "FailedGetResourceMetric",
						Message: "missing request for cpu",
					},
				}
				return hpa
			}(),
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusWarning,
				Details:    []component.Component{component.NewText("Scaling is not active: (FailedGetResourceMetric) missing request for cpu")},
				Properties: properties(2, 2),
			},
		},
		{
			name: "cannot scale",
			object: func() runtime.Object {
				hpa := newHPA()
				hpa.Status.CurrentReplicas = 2
				hpa.Status.DesiredReplicas = 2
				hpa.Status.Conditions = []autoscalingv2.HorizontalPodAutoscalerCondition{
					{
						Type:    autoscalingv2.AbleToScale,
						Status:  corev1.ConditionFalse,
						Reason:  "FailedGetScale",
						Message: "could not get scale",
					},
				}
				return hpa
			}(),
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusError,
				Details:    []component.Component{component.NewText("Cannot scale: (FailedGetScale) could not get scale")},
				Properties: properties(2, 2),
			},
		},
		{
			name: "scaling limited",
			object: func() runtime.Object {
				hpa := newHPA()
				hpa.Status.CurrentReplicas = 2
				hpa.Status.DesiredReplicas = 2
				hpa.Status.Conditions = []autoscalingv2.HorizontalPodAutoscalerCondition{
					{
						Type:    autoscalingv2.ScalingLimited,
						Status:  corev1.ConditionTrue,
						Reason:  "TooManyReplicas",
						Message: "at maximum",
					},
				}
				return hpa
			}(),
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusWarning,
				Details:    []component.Component{component.NewText("Scaling is limited: (TooManyReplicas) at maximum")},
				Properties: properties(2, 2),
			},
		},
		{
			name:   "nil object",
			object: nil,
			isErr:  true,
		},
		{
			name:   "not a horizontal pod autoscaler",
			object: &unstructured.Unstructured{},
			isErr:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := horizontalPodAutoscaler(context.Background(), tc.object, nil, nil)
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
	_ = autoscalingv2.AddToScheme(scheme.Scheme)
}
