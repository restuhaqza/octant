/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package overview

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/vmware-tanzu/octant/internal/gvk"
)

func Test_crdPath(t *testing.T) {
	got, err := crdPath("default", "crdName", "version", "name")
	require.NoError(t, err)

	expected := path.Join("/overview", "namespace", "default", "custom-resources", "crdName", "version", "name")
	assert.Equal(t, expected, got)
}

func Test_gvk_path(t *testing.T) {
	tests := []struct {
		name       string
		namespace  string
		apiVersion string
		kind       string
		objectName string
		expected   string
		isErr      bool
	}{
		{
			name:       "pod",
			namespace:  "default",
			apiVersion: "v1",
			kind:       "Pod",
			objectName: "pod",
			expected:   path.Join("/overview", "namespace", "default", "workloads", "pods", "pod"),
		},
		{
			name:       "horizontal pod autoscaler v1",
			namespace:  "default",
			apiVersion: "autoscaling/v1",
			kind:       "HorizontalPodAutoscaler",
			objectName: "hpa",
			expected:   path.Join("/overview", "namespace", "default", "discovery-and-load-balancing", "horizontal-pod-autoscalers", "hpa"),
		},
		{
			name:       "horizontal pod autoscaler v2",
			namespace:  "default",
			apiVersion: "autoscaling/v2",
			kind:       "HorizontalPodAutoscaler",
			objectName: "hpa",
			expected:   path.Join("/overview", "namespace", "default", "discovery-and-load-balancing", "horizontal-pod-autoscalers", "hpa"),
		},
		{
			name:       "gateway",
			namespace:  "default",
			apiVersion: "gateway.networking.k8s.io/v1",
			kind:       "Gateway",
			objectName: "gateway",
			expected:   path.Join("/overview", "namespace", "default", "gateway-api", "gateways", "gateway"),
		},
		{
			name:       "http route",
			namespace:  "default",
			apiVersion: "gateway.networking.k8s.io/v1",
			kind:       "HTTPRoute",
			objectName: "http-route",
			expected:   path.Join("/overview", "namespace", "default", "gateway-api", "httproutes", "http-route"),
		},
		{
			name:       "grpc route",
			namespace:  "default",
			apiVersion: "gateway.networking.k8s.io/v1",
			kind:       "GRPCRoute",
			objectName: "grpc-route",
			expected:   path.Join("/overview", "namespace", "default", "gateway-api", "grpcroutes", "grpc-route"),
		},
		{
			name:       "no namespace",
			apiVersion: "v1",
			kind:       "Pod",
			objectName: "pod",
			isErr:      true,
		},
		{
			name:       "unknown",
			namespace:  "default",
			apiVersion: "unknown",
			kind:       "ClusterRoleBinding",
			objectName: "cluster-role-binding",
			isErr:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := gvkPath(test.namespace, test.apiVersion, test.kind, test.objectName)
			if test.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, test.expected, got)
		})
	}
}

func Test_gvkReversePath(t *testing.T) {
	tests := []struct {
		name        string
		contentPath string
		expected    schema.GroupVersionKind
	}{
		{
			name:        "horizontal pod autoscaler",
			contentPath: "overview/namespace/default/discovery-and-load-balancing/horizontal-pod-autoscalers",
			expected:    gvk.HorizontalPodAutoscalerV2,
		},
		{
			name:        "gateway",
			contentPath: "overview/namespace/default/gateway-api/gateways",
			expected:    gvk.Gateway,
		},
		{
			name:        "http route",
			contentPath: "overview/namespace/default/gateway-api/httproutes",
			expected:    gvk.HTTPRoute,
		},
		{
			name:        "grpc route",
			contentPath: "overview/namespace/default/gateway-api/grpcroutes",
			expected:    gvk.GRPCRoute,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := gvkReversePath(test.contentPath, "default")
			require.NoError(t, err)
			assert.Equal(t, test.expected, got)
		})
	}
}
