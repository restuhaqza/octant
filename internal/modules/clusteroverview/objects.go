/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package clusteroverview

import (
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	corev1 "k8s.io/api/core/v1"
	flowcontrolv1 "k8s.io/api/flowcontrol/v1"
	networkingv1 "k8s.io/api/networking/v1"
	nodev1 "k8s.io/api/node/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	apiextv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiregistrationv1 "k8s.io/kube-aggregator/pkg/apis/apiregistration/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/vmware-tanzu/octant/internal/describer"
	"github.com/vmware-tanzu/octant/pkg/icon"
	"github.com/vmware-tanzu/octant/pkg/store"
)

var (
	customResourcesDescriber = describer.NewCRDSection(
		"/custom-resources",
		"Custom Resources",
	)

	crdsDescriber = describer.NewResource(describer.ResourceOptions{
		Path:           "/custom-resource-definitions",
		ObjectStoreKey: store.Key{APIVersion: "apiextensions.k8s.io/v1", Kind: "CustomResourceDefinition"},
		ListType:       &apiextv1.CustomResourceDefinitionList{},
		ObjectType:     &apiextv1.CustomResourceDefinition{},
		Titles:         describer.ResourceTitle{List: "Custom Resource Definitions", Object: "Custom Resource Definitions"},
		ClusterWide:    true,
		IconName:       icon.CustomResourceDefinition,
	})

	rbacClusterRoles = describer.NewResource(describer.ResourceOptions{
		Path:           "/rbac/cluster-roles",
		ObjectStoreKey: store.Key{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "ClusterRole"},
		ListType:       &rbacv1.ClusterRoleList{},
		ObjectType:     &rbacv1.ClusterRole{},
		Titles:         describer.ResourceTitle{List: "Cluster Roles", Object: "Cluster Roles"},
		ClusterWide:    true,
		IconName:       icon.ClusterOverviewClusterRole,
	})

	rbacClusterRoleBindings = describer.NewResource(describer.ResourceOptions{
		Path:           "/rbac/cluster-role-bindings",
		ObjectStoreKey: store.Key{APIVersion: "rbac.authorization.k8s.io/v1", Kind: "ClusterRoleBinding"},
		ListType:       &rbacv1.ClusterRoleBindingList{},
		ObjectType:     &rbacv1.ClusterRoleBinding{},
		Titles:         describer.ResourceTitle{List: "Cluster Role Bindings", Object: "Cluster Role Bindings"},
		ClusterWide:    true,
		IconName:       icon.ClusterOverviewClusterRoleBinding,
	})

	rbacDescriber = describer.NewSection(
		"/rbac",
		"RBAC",
		rbacClusterRoles,
		rbacClusterRoleBindings,
	)

	webhooksDescriber = describer.NewSection(
		"/webhooks",
		"Webhooks",
		webhooksMutatingWebhooks,
		webhooksValidatingWebhooks,
	)

	webhooksValidatingWebhooks = describer.NewResource(describer.ResourceOptions{
		Path:           "/webhooks/validating-webhooks",
		ObjectStoreKey: store.Key{APIVersion: "admissionregistration.k8s.io/v1", Kind: "ValidatingWebhookConfiguration"},
		ListType:       &admissionregistrationv1.ValidatingWebhookConfigurationList{},
		ObjectType:     &admissionregistrationv1.ValidatingWebhookConfiguration{},
		Titles:         describer.ResourceTitle{List: "Validating Webhooks", Object: "Validating Webhook Configuration"},
		ClusterWide:    true,
		IconName:       icon.Webhooks,
	})

	webhooksMutatingWebhooks = describer.NewResource(describer.ResourceOptions{
		Path:           "/webhooks/mutating-webhooks",
		ObjectStoreKey: store.Key{APIVersion: "admissionregistration.k8s.io/v1", Kind: "MutatingWebhookConfiguration"},
		ListType:       &admissionregistrationv1.MutatingWebhookConfigurationList{},
		ObjectType:     &admissionregistrationv1.MutatingWebhookConfiguration{},
		Titles:         describer.ResourceTitle{List: "Mutating Webhooks", Object: "Mutating Webhook Configuration"},
		ClusterWide:    true,
		IconName:       icon.Webhooks,
	})

	nodesDescriber = describer.NewResource(describer.ResourceOptions{
		Path:                  "/nodes",
		ObjectStoreKey:        store.Key{APIVersion: "v1", Kind: "Node"},
		ListType:              &corev1.NodeList{},
		ObjectType:            &corev1.Node{},
		Titles:                describer.ResourceTitle{List: "Nodes", Object: "Nodes"},
		DisableResourceViewer: true,
		ClusterWide:           true,
		IconName:              icon.ClusterOverviewNode,
	})

	storagePersistentVolumeDescriber = describer.NewResource(describer.ResourceOptions{
		Path:           "/storage/persistent-volumes",
		ObjectStoreKey: store.Key{APIVersion: "v1", Kind: "PersistentVolume"},
		ListType:       &corev1.PersistentVolumeList{},
		ObjectType:     &corev1.PersistentVolume{},
		Titles:         describer.ResourceTitle{List: "Persistent Volumes", Object: "Persistent Volumes"},
		ClusterWide:    true,
		IconName:       icon.ClusterOverviewPersistentVolume,
	})

	storageStorageClassDescriber = describer.NewResource(describer.ResourceOptions{
		Path:           "/storage/storage-classes",
		ObjectStoreKey: store.Key{APIVersion: "storage.k8s.io/v1", Kind: "StorageClass"},
		ListType:       &storagev1.StorageClassList{},
		ObjectType:     &storagev1.StorageClass{},
		Titles:         describer.ResourceTitle{List: "Storage Classes", Object: "Storage Classes"},
		ClusterWide:    true,
		IconName:       icon.ClusterOverviewStorageClass,
	})

	storageCSIDrivers = describer.NewResource(describer.ResourceOptions{
		Path:           "/storage/csi-drivers",
		ObjectStoreKey: store.Key{APIVersion: "storage.k8s.io/v1", Kind: "CSIDriver"},
		ListType:       &storagev1.CSIDriverList{},
		ObjectType:     &storagev1.CSIDriver{},
		Titles:         describer.ResourceTitle{List: "CSI Drivers", Object: "CSI Driver"},
		ClusterWide:    true,
		IconName:       icon.ConfigAndStorage,
	})

	storageCSINodes = describer.NewResource(describer.ResourceOptions{
		Path:           "/storage/csi-nodes",
		ObjectStoreKey: store.Key{APIVersion: "storage.k8s.io/v1", Kind: "CSINode"},
		ListType:       &storagev1.CSINodeList{},
		ObjectType:     &storagev1.CSINode{},
		Titles:         describer.ResourceTitle{List: "CSI Nodes", Object: "CSI Node"},
		ClusterWide:    true,
		IconName:       icon.ConfigAndStorage,
	})

	storageVolumeAttachments = describer.NewResource(describer.ResourceOptions{
		Path:           "/storage/volume-attachments",
		ObjectStoreKey: store.Key{APIVersion: "storage.k8s.io/v1", Kind: "VolumeAttachment"},
		ListType:       &storagev1.VolumeAttachmentList{},
		ObjectType:     &storagev1.VolumeAttachment{},
		Titles:         describer.ResourceTitle{List: "Volume Attachments", Object: "Volume Attachment"},
		ClusterWide:    true,
		IconName:       icon.ConfigAndStorage,
	})

	storageDescriber = describer.NewSection(
		"/storage",
		"Storage",
		storagePersistentVolumeDescriber,
		storageStorageClassDescriber,
		storageCSIDrivers,
		storageCSINodes,
		storageVolumeAttachments,
	)

	workloadsPriorityClasses = describer.NewResource(describer.ResourceOptions{
		Path:           "/workloads/priority-classes",
		ObjectStoreKey: store.Key{APIVersion: "scheduling.k8s.io/v1", Kind: "PriorityClass"},
		ListType:       &schedulingv1.PriorityClassList{},
		ObjectType:     &schedulingv1.PriorityClass{},
		Titles:         describer.ResourceTitle{List: "Priority Classes", Object: "Priority Class"},
		ClusterWide:    true,
		IconName:       icon.Workloads,
	})

	workloadsRuntimeClasses = describer.NewResource(describer.ResourceOptions{
		Path:           "/workloads/runtime-classes",
		ObjectStoreKey: store.Key{APIVersion: "node.k8s.io/v1", Kind: "RuntimeClass"},
		ListType:       &nodev1.RuntimeClassList{},
		ObjectType:     &nodev1.RuntimeClass{},
		Titles:         describer.ResourceTitle{List: "Runtime Classes", Object: "Runtime Class"},
		ClusterWide:    true,
		IconName:       icon.Workloads,
	})

	workloadsDescriber = describer.NewSection(
		"/workloads",
		"Workloads",
		workloadsPriorityClasses,
		workloadsRuntimeClasses,
	)

	dlbIngressClasses = describer.NewResource(describer.ResourceOptions{
		Path:           "/discovery-and-load-balancing/ingress-classes",
		ObjectStoreKey: store.Key{APIVersion: "networking.k8s.io/v1", Kind: "IngressClass"},
		ListType:       &networkingv1.IngressClassList{},
		ObjectType:     &networkingv1.IngressClass{},
		Titles:         describer.ResourceTitle{List: "Ingress Classes", Object: "Ingress Class"},
		ClusterWide:    true,
		IconName:       icon.DiscoveryAndLoadBalancing,
	})

	dlbDescriber = describer.NewSection(
		"/discovery-and-load-balancing",
		"Discovery and Load Balancing",
		dlbIngressClasses,
	)

	clusterFlowSchemas = describer.NewResource(describer.ResourceOptions{
		Path:           "/cluster/flow-schemas",
		ObjectStoreKey: store.Key{APIVersion: "flowcontrol.apiserver.k8s.io/v1", Kind: "FlowSchema"},
		ListType:       &flowcontrolv1.FlowSchemaList{},
		ObjectType:     &flowcontrolv1.FlowSchema{},
		Titles:         describer.ResourceTitle{List: "Flow Schemas", Object: "Flow Schema"},
		ClusterWide:    true,
		IconName:       icon.Cluster,
	})

	clusterPriorityLevelConfigurations = describer.NewResource(describer.ResourceOptions{
		Path:           "/cluster/priority-level-configurations",
		ObjectStoreKey: store.Key{APIVersion: "flowcontrol.apiserver.k8s.io/v1", Kind: "PriorityLevelConfiguration"},
		ListType:       &flowcontrolv1.PriorityLevelConfigurationList{},
		ObjectType:     &flowcontrolv1.PriorityLevelConfiguration{},
		Titles:         describer.ResourceTitle{List: "Priority Level Configurations", Object: "Priority Level Configuration"},
		ClusterWide:    true,
		IconName:       icon.Cluster,
	})

	clusterValidatingAdmissionPolicies = describer.NewResource(describer.ResourceOptions{
		Path:           "/cluster/validating-admission-policies",
		ObjectStoreKey: store.Key{APIVersion: "admissionregistration.k8s.io/v1", Kind: "ValidatingAdmissionPolicy"},
		ListType:       &admissionregistrationv1.ValidatingAdmissionPolicyList{},
		ObjectType:     &admissionregistrationv1.ValidatingAdmissionPolicy{},
		Titles:         describer.ResourceTitle{List: "Validating Admission Policies", Object: "Validating Admission Policy"},
		ClusterWide:    true,
		IconName:       icon.Webhooks,
	})

	clusterDescriber = describer.NewSection(
		"/cluster",
		"Cluster",
		clusterFlowSchemas,
		clusterPriorityLevelConfigurations,
		clusterValidatingAdmissionPolicies,
	)

	namespacesDescriber = describer.NewResource(describer.ResourceOptions{
		Path:                  "/namespaces",
		ObjectStoreKey:        store.Key{APIVersion: "v1", Kind: "Namespace"},
		ListType:              &corev1.NamespaceList{},
		ObjectType:            &corev1.Namespace{},
		Titles:                describer.ResourceTitle{List: "Namespaces", Object: "Namespaces"},
		DisableResourceViewer: true,
		ClusterWide:           true,
		IconName:              icon.ClusterOverviewNamespace,
	})

	portForwardDescriber = NewPortForwardListDescriber()

	apiServerDescriber = describer.NewSection(
		"/api-server",
		"API Server",
		apiServerApiServices,
	)

	apiServerApiServices = describer.NewResource(describer.ResourceOptions{
		Path:           "/api-server/api-services",
		ObjectStoreKey: store.Key{APIVersion: "apiregistration.k8s.io/v1", Kind: "APIService"},
		ListType:       &apiregistrationv1.APIServiceList{},
		ObjectType:     &apiregistrationv1.APIService{},
		Titles:         describer.ResourceTitle{List: "API Services", Object: "API Services"},
		ClusterWide:    true,
		IconName:       icon.ApiServer,
	})

	gatewayAPIGatewayClasses = describer.NewResource(describer.ResourceOptions{
		Path:           "/gateway-api/gateway-classes",
		ObjectStoreKey: store.Key{APIVersion: "gateway.networking.k8s.io/v1", Kind: "GatewayClass"},
		ListType:       &gatewayv1.GatewayClassList{},
		ObjectType:     &gatewayv1.GatewayClass{},
		Titles:         describer.ResourceTitle{List: "Gateway Classes", Object: "Gateway Classes"},
		ClusterWide:    true,
		IconName:       icon.DiscoveryAndLoadBalancing,
	})

	gatewayAPIDescriber = describer.NewSection(
		"/gateway-api",
		"Gateway API",
		gatewayAPIGatewayClasses,
	)

	rootDescriber = describer.NewSection(
		"/",
		"Cluster Overview",
		namespacesDescriber,
		customResourcesDescriber,
		crdsDescriber,
		rbacDescriber,
		webhooksDescriber,
		workloadsDescriber,
		dlbDescriber,
		clusterDescriber,
		nodesDescriber,
		storageDescriber,
		portForwardDescriber,
		apiServerDescriber,
		gatewayAPIDescriber,
	)
)
