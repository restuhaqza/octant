/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package clusteroverview

import (
	"path"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_crdPath(t *testing.T) {
	got, err := crdPath("namespace", "crdName", "version", "name")
	require.NoError(t, err)

	expected := path.Join("/cluster-overview", "custom-resources", "crdName", "version", "name")
	assert.Equal(t, expected, got)
}

func Test_gvk_path(t *testing.T) {
	tests := []struct {
		name       string
		apiVersion string
		kind       string
		objectName string
		expected   string
		isErr      bool
	}{
		{
			name:       "ClusterRole",
			apiVersion: rbacAPIVersion,
			kind:       "ClusterRole",
			objectName: "cluster-role",
			expected:   path.Join("/cluster-overview", "rbac", "cluster-roles", "cluster-role"),
		},
		{
			name:       "ClusterRoleBinding",
			apiVersion: rbacAPIVersion,
			kind:       "ClusterRoleBinding",
			objectName: "cluster-role-binding",
			expected:   path.Join("/cluster-overview", "rbac", "cluster-role-bindings", "cluster-role-binding"),
		},
		{
			name:       "PriorityClass",
			apiVersion: "scheduling.k8s.io/v1",
			kind:       "PriorityClass",
			objectName: "priority",
			expected:   path.Join("/cluster-overview", "workloads", "priority-classes", "priority"),
		},
		{
			name:       "RuntimeClass",
			apiVersion: "node.k8s.io/v1",
			kind:       "RuntimeClass",
			objectName: "runtime",
			expected:   path.Join("/cluster-overview", "workloads", "runtime-classes", "runtime"),
		},
		{
			name:       "IngressClass",
			apiVersion: "networking.k8s.io/v1",
			kind:       "IngressClass",
			objectName: "ingress",
			expected:   path.Join("/cluster-overview", "discovery-and-load-balancing", "ingress-classes", "ingress"),
		},
		{
			name:       "CSIDriver",
			apiVersion: "storage.k8s.io/v1",
			kind:       "CSIDriver",
			objectName: "driver",
			expected:   path.Join("/cluster-overview", "storage", "csi-drivers", "driver"),
		},
		{
			name:       "CSINode",
			apiVersion: "storage.k8s.io/v1",
			kind:       "CSINode",
			objectName: "csi-node",
			expected:   path.Join("/cluster-overview", "storage", "csi-nodes", "csi-node"),
		},
		{
			name:       "VolumeAttachment",
			apiVersion: "storage.k8s.io/v1",
			kind:       "VolumeAttachment",
			objectName: "attachment",
			expected:   path.Join("/cluster-overview", "storage", "volume-attachments", "attachment"),
		},
		{
			name:       "FlowSchema",
			apiVersion: "flowcontrol.apiserver.k8s.io/v1",
			kind:       "FlowSchema",
			objectName: "flow",
			expected:   path.Join("/cluster-overview", "cluster", "flow-schemas", "flow"),
		},
		{
			name:       "PriorityLevelConfiguration",
			apiVersion: "flowcontrol.apiserver.k8s.io/v1",
			kind:       "PriorityLevelConfiguration",
			objectName: "plc",
			expected:   path.Join("/cluster-overview", "cluster", "priority-level-configurations", "plc"),
		},
		{
			name:       "ValidatingAdmissionPolicy",
			apiVersion: "admissionregistration.k8s.io/v1",
			kind:       "ValidatingAdmissionPolicy",
			objectName: "vap",
			expected:   path.Join("/cluster-overview", "cluster", "validating-admission-policies", "vap"),
		},
		{
			name:       "unknown",
			apiVersion: "unknown",
			kind:       "ClusterRoleBinding",
			objectName: "cluster-role-binding",
			isErr:      true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := gvkPath("", test.apiVersion, test.kind, test.objectName)
			if test.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, test.expected, got)
		})
	}
}
