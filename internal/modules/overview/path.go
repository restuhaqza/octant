/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package overview

import (
	"path"
	"strings"

	"github.com/vmware-tanzu/octant/internal/util/path_util"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/vmware-tanzu/octant/internal/gvk"
)

var (
	supportedGVKs = []schema.GroupVersionKind{
		gvk.AppReplicaSet,
		gvk.CronJob,
		gvk.CSIStorageCapacity,
		gvk.DaemonSet,
		gvk.Deployment,
		gvk.Job,
		gvk.Lease,
		gvk.LimitRange,
		gvk.Pod,
		gvk.PodDisruptionBudget,
		gvk.ReplicationController,
		gvk.ResourceQuota,
		gvk.StatefulSet,
		gvk.HorizontalPodAutoscaler,
		gvk.HorizontalPodAutoscalerV2,
		gvk.Gateway,
		gvk.HTTPRoute,
		gvk.GRPCRoute,
		gvk.Ingress,
		gvk.Service,
		gvk.NetworkPolicy,
		gvk.ConfigMap,
		gvk.Secret,
		gvk.PersistentVolumeClaim,
		gvk.ServiceAccount,
		gvk.RoleBinding,
		gvk.Role,
		gvk.Event,
	}
)

func crdPath(namespace, crdName, version, name string) (string, error) {
	if namespace == "" {
		return "", errors.Errorf("unable to create CRD path for %s due to missing namespace", crdName)
	}

	return path_util.NamespacedPath("/overview", namespace, "custom-resources", crdName, version, name), nil
}

func gvkPath(namespace, apiVersion, kind, name string) (string, error) {
	if namespace == "" {
		return "", errors.Errorf("unable to create  path for %s %s due to missing namespace", apiVersion, kind)
	}

	var p string

	switch {
	case apiVersion == "apps/v1" && kind == "DaemonSet":
		p = "/workloads/daemon-sets"
	case apiVersion == "apps/v1" && kind == "ReplicaSet":
		p = "/workloads/replica-sets"
	case apiVersion == "apps/v1" && kind == "StatefulSet":
		p = "/workloads/stateful-sets"
	case apiVersion == "apps/v1" && kind == "Deployment":
		p = "/workloads/deployments"
	case apiVersion == "batch/v1" && kind == "CronJob":
		p = "/workloads/cron-jobs"
	case apiVersion == "batch/v1" && kind == "Job":
		p = "/workloads/jobs"
	case apiVersion == "v1" && kind == "ReplicationController":
		p = "/workloads/replication-controllers"
	case apiVersion == "v1" && kind == "Secret":
		p = "/config-and-storage/secrets"
	case apiVersion == "v1" && kind == "ConfigMap":
		p = "/config-and-storage/config-maps"
	case apiVersion == "v1" && kind == "PersistentVolumeClaim":
		p = "/config-and-storage/persistent-volume-claims"
	case apiVersion == "v1" && kind == "ServiceAccount":
		p = "/config-and-storage/service-accounts"
	case (apiVersion == "autoscaling/v1" || apiVersion == "autoscaling/v2") && kind == "HorizontalPodAutoscaler":
		p = "/discovery-and-load-balancing/horizontal-pod-autoscalers"
	case apiVersion == "gateway.networking.k8s.io/v1" && kind == "Gateway":
		p = "/gateway-api/gateways"
	case apiVersion == "gateway.networking.k8s.io/v1" && kind == "HTTPRoute":
		p = "/gateway-api/httproutes"
	case apiVersion == "gateway.networking.k8s.io/v1" && kind == "GRPCRoute":
		p = "/gateway-api/grpcroutes"
	case apiVersion == "networking.k8s.io/v1" && kind == "Ingress":
		p = "/discovery-and-load-balancing/ingresses"
	case apiVersion == "v1" && kind == "Service":
		p = "/discovery-and-load-balancing/services"
	case apiVersion == "networking.k8s.io/v1" && kind == "NetworkPolicy":
		p = "/discovery-and-load-balancing/network-policies"
	case apiVersion == "rbac.authorization.k8s.io/v1" && kind == "Role":
		p = "/rbac/roles"
	case apiVersion == "rbac.authorization.k8s.io/v1" && kind == "RoleBinding":
		p = "/rbac/role-bindings"
	case apiVersion == "v1" && kind == "Event":
		p = "/events"
	case apiVersion == "v1" && kind == "Pod":
		p = "/workloads/pods"
	case apiVersion == "policy/v1" && kind == "PodDisruptionBudget":
		p = "/workloads/pod-disruption-budgets"
	case apiVersion == "v1" && kind == "ResourceQuota":
		p = "/config-and-storage/resource-quotas"
	case apiVersion == "v1" && kind == "LimitRange":
		p = "/config-and-storage/limit-ranges"
	case apiVersion == "coordination.k8s.io/v1" && kind == "Lease":
		p = "/config-and-storage/leases"
	case apiVersion == "storage.k8s.io/v1" && kind == "CSIStorageCapacity":
		p = "/config-and-storage/csi-storage-capacities"
	default:
		return "", errors.Errorf("unknown object %s %s", apiVersion, kind)
	}

	return path.Join("/overview/namespace", namespace, p, name), nil
}

func gvkReversePath(contentPath, namespace string) (schema.GroupVersionKind, error) {
	reducedPath := strings.Replace(contentPath, path.Join("overview/namespace", namespace), "", 1)

	switch {
	case reducedPath == "/workloads/daemon-sets":
		return gvk.DaemonSet, nil
	case reducedPath == "/workloads/replica-sets":
		return gvk.AppReplicaSet, nil
	case reducedPath == "/workloads/stateful-sets":
		return gvk.StatefulSet, nil
	case reducedPath == "/workloads/deployments":
		return gvk.Deployment, nil
	case reducedPath == "/workloads/cron-jobs":
		return gvk.CronJob, nil
	case reducedPath == "/workloads/jobs":
		return gvk.Job, nil
	case reducedPath == "/workloads/replication-controllers":
		return gvk.ReplicationController, nil
	case reducedPath == "/config-and-storage/secrets":
		return gvk.Secret, nil
	case reducedPath == "/config-and-storage/config-maps":
		return gvk.ConfigMap, nil
	case reducedPath == "/config-and-storage/persistent-volume-claims":
		return gvk.PersistentVolumeClaim, nil
	case reducedPath == "/config-and-storage/service-accounts":
		return gvk.ServiceAccount, nil
	case reducedPath == "/discovery-and-load-balancing/horizontal-pod-autoscalers":
		return gvk.HorizontalPodAutoscalerV2, nil
	case reducedPath == "/gateway-api/gateways":
		return gvk.Gateway, nil
	case reducedPath == "/gateway-api/httproutes":
		return gvk.HTTPRoute, nil
	case reducedPath == "/gateway-api/grpcroutes":
		return gvk.GRPCRoute, nil
	case reducedPath == "/discovery-and-load-balancing/ingresses":
		return gvk.Ingress, nil
	case reducedPath == "/discovery-and-load-balancing/services":
		return gvk.Service, nil
	case reducedPath == "/discovery-and-load-balancing/network-policies":
		return gvk.NetworkPolicy, nil
	case reducedPath == "/rbac/roles":
		return gvk.Role, nil
	case reducedPath == "/rbac/role-bindings":
		return gvk.RoleBinding, nil
	case reducedPath == "/events":
		return gvk.Event, nil
	case reducedPath == "/workloads/pods":
		return gvk.Pod, nil
	case reducedPath == "/workloads/pod-disruption-budgets":
		return gvk.PodDisruptionBudget, nil
	case reducedPath == "/config-and-storage/resource-quotas":
		return gvk.ResourceQuota, nil
	case reducedPath == "/config-and-storage/limit-ranges":
		return gvk.LimitRange, nil
	case reducedPath == "/config-and-storage/leases":
		return gvk.Lease, nil
	case reducedPath == "/config-and-storage/csi-storage-capacities":
		return gvk.CSIStorageCapacity, nil
	default:
		return schema.GroupVersionKind{}, errors.Errorf("unknown gvk %s", contentPath)
	}
}
