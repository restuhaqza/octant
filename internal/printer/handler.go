/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

// Handler configures handlers for a printer.
type Handler interface {
	Handler(printFunc interface{}) error
}

// AddHandlers adds print handlers to a printer.
func AddHandlers(p Handler) error {
	handlers := []interface{}{
		EventListHandler,
		EventHandler,
		ClusterRoleBindingListHandler,
		ClusterRoleBindingHandler,
		ConfigMapListHandler,
		ConfigMapHandler,
		CronJobListHandler,
		CronJobHandler,
		ClusterRoleListHandler,
		ClusterRoleHandler,
		CSIDriverListHandler,
		CSIDriverHandler,
		CSINodeListHandler,
		CSINodeHandler,
		CSIStorageCapacityListHandler,
		CSIStorageCapacityHandler,
		CustomResourceDefinitionListHandler,
		CustomResourceDefinitionHandler,
		DaemonSetListHandler,
		DaemonSetHandler,
		DeploymentHandler,
		DeploymentListHandler,
		FlowSchemaListHandler,
		FlowSchemaHandler,
		HorizontalPodAutoscalerHandler,
		HorizontalPodAutoscalerListHandler,
		HorizontalPodAutoscalerV2Handler,
		HorizontalPodAutoscalerV2ListHandler,
		IngressListHandler,
		IngressHandler,
		IngressClassListHandler,
		IngressClassHandler,
		JobListHandler,
		JobHandler,
		LeaseListHandler,
		LeaseHandler,
		LimitRangeListHandler,
		LimitRangeHandler,
		NodeHandler,
		NodeListHandler,
		NamespaceHandler,
		NamespaceListHandler,
		NetworkPolicyHandler,
		NetworkPolicyListHandler,
		PodDisruptionBudgetListHandler,
		PodDisruptionBudgetHandler,
		PriorityClassListHandler,
		PriorityClassHandler,
		PriorityLevelConfigurationListHandler,
		PriorityLevelConfigurationHandler,
		ResourceQuotaListHandler,
		ResourceQuotaHandler,
		RuntimeClassListHandler,
		RuntimeClassHandler,
		ReplicaSetHandler,
		ReplicaSetListHandler,
		ReplicationControllerHandler,
		ReplicationControllerListHandler,
		PodHandler,
		PodListHandler,
		PersistentVolumeHandler,
		PersistentVolumeListHandler,
		PersistentVolumeClaimHandler,
		PersistentVolumeClaimListHandler,
		ServiceAccountListHandler,
		ServiceAccountHandler,
		ServiceHandler,
		ServiceListHandler,
		SecretHandler,
		SecretListHandler,
		StatefulSetHandler,
		StatefulSetListHandler,
		StorageClassHandler,
		StorageClassListHandler,
		ValidatingAdmissionPolicyListHandler,
		ValidatingAdmissionPolicyHandler,
		VolumeAttachmentListHandler,
		VolumeAttachmentHandler,
		RoleBindingListHandler,
		RoleBindingHandler,
		RoleListHandler,
		RoleHandler,
		APIServiceHandler,
		APIServiceListHandler,
		MutatingWebhookConfigurationHandler,
		MutatingWebhookConfigurationListHandler,
		ValidatingWebhookConfigurationHandler,
		ValidatingWebhookConfigurationListHandler,
	}

	for _, handler := range handlers {
		if err := p.Handler(handler); err != nil {
			return err
		}
	}

	return nil
}
