/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectstatus

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	linkFake "github.com/vmware-tanzu/octant/internal/link/fake"
	"github.com/vmware-tanzu/octant/internal/testutil"
	storefake "github.com/vmware-tanzu/octant/pkg/store/fake"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

func Test_podDisruptionBudget(t *testing.T) {
	cases := []struct {
		name     string
		init     func(*testing.T) runtime.Object
		expected ObjectStatus
		isErr    bool
	}{
		{
			name: "healthy",
			init: func(t *testing.T) runtime.Object {
				pdb := testutil.CreatePodDisruptionBudget("pdb")
				pdb.Status.CurrentHealthy = 2
				pdb.Status.DesiredHealthy = 2
				pdb.Status.DisruptionsAllowed = 1
				return pdb
			},
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusOK,
				Details: []component.Component{
					component.NewText("1 disruptions allowed"),
				},
				Properties: nil,
			},
		},
		{
			name: "unhealthy",
			init: func(t *testing.T) runtime.Object {
				pdb := testutil.CreatePodDisruptionBudget("pdb")
				pdb.Status.CurrentHealthy = 1
				pdb.Status.DesiredHealthy = 2
				return pdb
			},
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusWarning,
				Details: []component.Component{
					component.NewText("1 current healthy, 2 desired healthy"),
				},
				Properties: nil,
			},
		},
		{
			name: "object is nil",
			init: func(t *testing.T) runtime.Object {
				return nil
			},
			isErr: true,
		},
		{
			name: "object is not a pod disruption budget",
			init: func(t *testing.T) runtime.Object {
				return &unstructured.Unstructured{}
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			linkInterface := linkFake.NewMockInterface(controller)
			defer controller.Finish()

			o := storefake.NewMockStore(controller)
			object := tc.init(t)
			ctx := context.Background()

			status, err := podDisruptionBudget(ctx, object, o, linkInterface)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tc.expected, status)
		})
	}
}

func Test_resourceQuota(t *testing.T) {
	cases := []struct {
		name     string
		init     func(*testing.T) runtime.Object
		expected ObjectStatus
		isErr    bool
	}{
		{
			name: "within quota",
			init: func(t *testing.T) runtime.Object {
				quota := testutil.CreateResourceQuota("quota")
				quota.Status.Hard = corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("2"),
				}
				quota.Status.Used = corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("1"),
				}
				return quota
			},
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusOK,
				Details: []component.Component{
					component.NewText("v1 ResourceQuota is OK"),
				},
				Properties: nil,
			},
		},
		{
			name: "exceeded quota",
			init: func(t *testing.T) runtime.Object {
				quota := testutil.CreateResourceQuota("quota")
				quota.Status.Hard = corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("1"),
				}
				quota.Status.Used = corev1.ResourceList{
					corev1.ResourceCPU: resource.MustParse("2"),
				}
				return quota
			},
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusError,
				Details: []component.Component{
					component.NewText("Exceeded quota: cpu"),
				},
				Properties: nil,
			},
		},
		{
			name: "object is nil",
			init: func(t *testing.T) runtime.Object {
				return nil
			},
			isErr: true,
		},
		{
			name: "object is not a resource quota",
			init: func(t *testing.T) runtime.Object {
				return &unstructured.Unstructured{}
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			linkInterface := linkFake.NewMockInterface(controller)
			defer controller.Finish()

			o := storefake.NewMockStore(controller)
			object := tc.init(t)
			ctx := context.Background()

			status, err := resourceQuota(ctx, object, o, linkInterface)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tc.expected, status)
		})
	}
}
