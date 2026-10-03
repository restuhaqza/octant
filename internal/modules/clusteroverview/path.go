/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package clusteroverview

import (
	"fmt"
	"path"

	"github.com/pkg/errors"

	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/vmware-tanzu/octant/internal/gvk"
)

var (
	supportedGVKs = []schema.GroupVersionKind{
		gvk.ClusterRoleBinding,
		gvk.ClusterRole,
		gvk.CSIDriver,
		gvk.CSINode,
		gvk.FlowSchema,
		gvk.IngressClass,
		gvk.Node,
		gvk.PersistentVolume,
		gvk.Namespace,
		gvk.PriorityClass,
		gvk.PriorityLevelConfiguration,
		gvk.RuntimeClass,
		gvk.CustomResourceDefinition,
		gvk.APIService,
		gvk.MutatingWebhookConfiguration,
		gvk.ValidatingWebhookConfiguration,
		gvk.ValidatingAdmissionPolicy,
		gvk.VolumeAttachment,
		gvk.StorageClass,
		gvk.GatewayClass,
	}
)

const rbacAPIVersion = "rbac.authorization.k8s.io/v1"

func crdPath(namespace, crdName, version, name string) (string, error) {
	return path.Join("/cluster-overview/custom-resources", crdName, version, name), nil
}

func gvkPath(namespace, apiVersion, kind, name string) (string, error) {
	var p string

	switch {
	case apiVersion == rbacAPIVersion && kind == "ClusterRole":
		p = "/rbac/cluster-roles"
	case apiVersion == rbacAPIVersion && kind == "ClusterRoleBinding":
		p = "/rbac/cluster-role-bindings"
	case apiVersion == "v1" && kind == "Node":
		p = "/nodes"
	case apiVersion == "v1" && kind == "PersistentVolume":
		p = "/storage/persistent-volumes"
	case apiVersion == "v1" && kind == "Namespace":
		p = "/namespaces"
	case apiVersion == gvk.CustomResourceDefinition.GroupVersion().String() &&
		kind == gvk.CustomResourceDefinition.Kind:
		p = "/custom-resource-definitions"
	case apiVersion == "apiregistration.k8s.io/v1" && kind == "APIService":
		p = "/api-server/api-services"
	case apiVersion == "admissionregistration.k8s.io/v1" && kind == "MutatingWebhookConfiguration":
		p = "/webhooks/mutating-webhooks"
	case apiVersion == "admissionregistration.k8s.io/v1" && kind == "ValidatingWebhookConfiguration":
		p = "/webhooks/validating-webhooks"
	case apiVersion == "storage.k8s.io/v1" && kind == "StorageClass":
		p = "/storage/storage-classes"
	case apiVersion == "storage.k8s.io/v1" && kind == "CSIDriver":
		p = "/storage/csi-drivers"
	case apiVersion == "storage.k8s.io/v1" && kind == "CSINode":
		p = "/storage/csi-nodes"
	case apiVersion == "storage.k8s.io/v1" && kind == "VolumeAttachment":
		p = "/storage/volume-attachments"
	case apiVersion == "scheduling.k8s.io/v1" && kind == "PriorityClass":
		p = "/workloads/priority-classes"
	case apiVersion == "node.k8s.io/v1" && kind == "RuntimeClass":
		p = "/workloads/runtime-classes"
	case apiVersion == "networking.k8s.io/v1" && kind == "IngressClass":
		p = "/discovery-and-load-balancing/ingress-classes"
	case apiVersion == "flowcontrol.apiserver.k8s.io/v1" && kind == "FlowSchema":
		p = "/cluster/flow-schemas"
	case apiVersion == "flowcontrol.apiserver.k8s.io/v1" && kind == "PriorityLevelConfiguration":
		p = "/cluster/priority-level-configurations"
	case apiVersion == "admissionregistration.k8s.io/v1" && kind == "ValidatingAdmissionPolicy":
		p = "/cluster/validating-admission-policies"
	case apiVersion == "gateway.networking.k8s.io/v1" && kind == "GatewayClass":
		p = "/gateway-api/gateway-classes"
	default:
		return "", fmt.Errorf("unknown object %s %s", apiVersion, kind)
	}

	return path.Join("/cluster-overview", p, name), nil
}

func gvkReversePath(contentPath, _ string) (schema.GroupVersionKind, error) {

	switch {
	case contentPath == "cluster-overview/rbac/cluster-roles":
		return gvk.ClusterRole, nil
	case contentPath == "cluster-overview/rbac/cluster-role-bindings":
		return gvk.ClusterRoleBinding, nil
	case contentPath == "cluster-overview/nodes":
		return gvk.Node, nil
	case contentPath == "cluster-overview/storage/persistent-volumes":
		return gvk.PersistentVolume, nil
	case contentPath == "cluster-overview/namespaces":
		return gvk.Namespace, nil
	case contentPath == "cluster-overview/custom-resource-definitions":
		return gvk.CustomResourceDefinition, nil
	case contentPath == "cluster-overview/api-server/api-services":
		return gvk.APIService, nil
	case contentPath == "cluster-overview/webhooks/mutating-webhooks":
		return gvk.MutatingWebhookConfiguration, nil
	case contentPath == "cluster-overview/webhooks/validating-webhooks":
		return gvk.ValidatingWebhookConfiguration, nil
	case contentPath == "cluster-overview/storage/storage-classes":
		return gvk.StorageClass, nil
	case contentPath == "cluster-overview/storage/csi-drivers":
		return gvk.CSIDriver, nil
	case contentPath == "cluster-overview/storage/csi-nodes":
		return gvk.CSINode, nil
	case contentPath == "cluster-overview/storage/volume-attachments":
		return gvk.VolumeAttachment, nil
	case contentPath == "cluster-overview/workloads/priority-classes":
		return gvk.PriorityClass, nil
	case contentPath == "cluster-overview/workloads/runtime-classes":
		return gvk.RuntimeClass, nil
	case contentPath == "cluster-overview/discovery-and-load-balancing/ingress-classes":
		return gvk.IngressClass, nil
	case contentPath == "cluster-overview/cluster/flow-schemas":
		return gvk.FlowSchema, nil
	case contentPath == "cluster-overview/cluster/priority-level-configurations":
		return gvk.PriorityLevelConfiguration, nil
	case contentPath == "cluster-overview/cluster/validating-admission-policies":
		return gvk.ValidatingAdmissionPolicy, nil
	case contentPath == "cluster-overview/gateway-api/gateway-classes":
		return gvk.GatewayClass, nil
	default:
		return schema.GroupVersionKind{}, errors.Errorf("unknown gvk %s", contentPath)
	}
}
