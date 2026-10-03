/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

// Package upgradescanner provides a read-only scan that finds live Kubernetes
// objects using API versions that have been removed or deprecated in a target
// Kubernetes version. Nothing in this package writes to the cluster.
package upgradescanner

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// APIRemoval describes an API version/kind that is removed in a Kubernetes
// release, along with the replacement API.
type APIRemoval struct {
	Group       string // e.g. "batch"; "" for the core group
	Version     string // e.g. "v1beta1"
	Kind        string // e.g. "CronJob"
	RemovedIn   string // e.g. "1.25"
	Replacement string // human readable replacement, e.g. "batch/v1 CronJob"; "none (removed)" when there is none

	// ReplaceAPIVersion and ReplaceKind describe the replacement API in a form
	// that can be used to generate an object link. They are empty when the API
	// has no replacement.
	ReplaceAPIVersion string
	ReplaceKind       string
}

// Removals is the deprecation database used by the scanner. It intentionally
// covers the well known removals called out by the Kubernetes deprecated API
// migration guide.
var Removals = []APIRemoval{
	// extensions/v1beta1.
	{Group: "extensions", Version: "v1beta1", Kind: "Deployment", RemovedIn: "1.16", Replacement: "apps/v1 Deployment", ReplaceAPIVersion: "apps/v1", ReplaceKind: "Deployment"},
	{Group: "extensions", Version: "v1beta1", Kind: "DaemonSet", RemovedIn: "1.16", Replacement: "apps/v1 DaemonSet", ReplaceAPIVersion: "apps/v1", ReplaceKind: "DaemonSet"},
	{Group: "extensions", Version: "v1beta1", Kind: "ReplicaSet", RemovedIn: "1.16", Replacement: "apps/v1 ReplicaSet", ReplaceAPIVersion: "apps/v1", ReplaceKind: "ReplicaSet"},
	{Group: "extensions", Version: "v1beta1", Kind: "NetworkPolicy", RemovedIn: "1.16", Replacement: "networking.k8s.io/v1 NetworkPolicy", ReplaceAPIVersion: "networking.k8s.io/v1", ReplaceKind: "NetworkPolicy"},
	{Group: "extensions", Version: "v1beta1", Kind: "Ingress", RemovedIn: "1.22", Replacement: "networking.k8s.io/v1 Ingress", ReplaceAPIVersion: "networking.k8s.io/v1", ReplaceKind: "Ingress"},
	{Group: "extensions", Version: "v1beta1", Kind: "PodSecurityPolicy", RemovedIn: "1.16", Replacement: "none (removed)"},

	// apps/v1beta1.
	{Group: "apps", Version: "v1beta1", Kind: "Deployment", RemovedIn: "1.16", Replacement: "apps/v1 Deployment", ReplaceAPIVersion: "apps/v1", ReplaceKind: "Deployment"},
	{Group: "apps", Version: "v1beta1", Kind: "StatefulSet", RemovedIn: "1.16", Replacement: "apps/v1 StatefulSet", ReplaceAPIVersion: "apps/v1", ReplaceKind: "StatefulSet"},
	{Group: "apps", Version: "v1beta1", Kind: "ControllerRevision", RemovedIn: "1.16", Replacement: "apps/v1 ControllerRevision", ReplaceAPIVersion: "apps/v1", ReplaceKind: "ControllerRevision"},

	// apps/v1beta2 (removed in 1.16).
	{Group: "apps", Version: "v1beta2", Kind: "Deployment", RemovedIn: "1.16", Replacement: "apps/v1 Deployment", ReplaceAPIVersion: "apps/v1", ReplaceKind: "Deployment"},
	{Group: "apps", Version: "v1beta2", Kind: "DaemonSet", RemovedIn: "1.16", Replacement: "apps/v1 DaemonSet", ReplaceAPIVersion: "apps/v1", ReplaceKind: "DaemonSet"},
	{Group: "apps", Version: "v1beta2", Kind: "ReplicaSet", RemovedIn: "1.16", Replacement: "apps/v1 ReplicaSet", ReplaceAPIVersion: "apps/v1", ReplaceKind: "ReplicaSet"},
	{Group: "apps", Version: "v1beta2", Kind: "StatefulSet", RemovedIn: "1.16", Replacement: "apps/v1 StatefulSet", ReplaceAPIVersion: "apps/v1", ReplaceKind: "StatefulSet"},
	{Group: "apps", Version: "v1beta2", Kind: "ControllerRevision", RemovedIn: "1.16", Replacement: "apps/v1 ControllerRevision", ReplaceAPIVersion: "apps/v1", ReplaceKind: "ControllerRevision"},

	// apiextensions.k8s.io/v1beta1 (removed in 1.22).
	{Group: "apiextensions.k8s.io", Version: "v1beta1", Kind: "CustomResourceDefinition", RemovedIn: "1.22", Replacement: "apiextensions.k8s.io/v1 CustomResourceDefinition", ReplaceAPIVersion: "apiextensions.k8s.io/v1", ReplaceKind: "CustomResourceDefinition"},

	// admissionregistration.k8s.io/v1beta1 (removed in 1.22).
	{Group: "admissionregistration.k8s.io", Version: "v1beta1", Kind: "MutatingWebhookConfiguration", RemovedIn: "1.22", Replacement: "admissionregistration.k8s.io/v1 MutatingWebhookConfiguration", ReplaceAPIVersion: "admissionregistration.k8s.io/v1", ReplaceKind: "MutatingWebhookConfiguration"},
	{Group: "admissionregistration.k8s.io", Version: "v1beta1", Kind: "ValidatingWebhookConfiguration", RemovedIn: "1.22", Replacement: "admissionregistration.k8s.io/v1 ValidatingWebhookConfiguration", ReplaceAPIVersion: "admissionregistration.k8s.io/v1", ReplaceKind: "ValidatingWebhookConfiguration"},

	// networking.k8s.io/v1beta1 (removed in 1.22).
	{Group: "networking.k8s.io", Version: "v1beta1", Kind: "Ingress", RemovedIn: "1.22", Replacement: "networking.k8s.io/v1 Ingress", ReplaceAPIVersion: "networking.k8s.io/v1", ReplaceKind: "Ingress"},
	{Group: "networking.k8s.io", Version: "v1beta1", Kind: "IngressClass", RemovedIn: "1.22", Replacement: "networking.k8s.io/v1 IngressClass", ReplaceAPIVersion: "networking.k8s.io/v1", ReplaceKind: "IngressClass"},

	// rbac.authorization.k8s.io/v1beta1 (removed in 1.22).
	{Group: "rbac.authorization.k8s.io", Version: "v1beta1", Kind: "ClusterRole", RemovedIn: "1.22", Replacement: "rbac.authorization.k8s.io/v1 ClusterRole", ReplaceAPIVersion: "rbac.authorization.k8s.io/v1", ReplaceKind: "ClusterRole"},
	{Group: "rbac.authorization.k8s.io", Version: "v1beta1", Kind: "ClusterRoleBinding", RemovedIn: "1.22", Replacement: "rbac.authorization.k8s.io/v1 ClusterRoleBinding", ReplaceAPIVersion: "rbac.authorization.k8s.io/v1", ReplaceKind: "ClusterRoleBinding"},
	{Group: "rbac.authorization.k8s.io", Version: "v1beta1", Kind: "Role", RemovedIn: "1.22", Replacement: "rbac.authorization.k8s.io/v1 Role", ReplaceAPIVersion: "rbac.authorization.k8s.io/v1", ReplaceKind: "Role"},
	{Group: "rbac.authorization.k8s.io", Version: "v1beta1", Kind: "RoleBinding", RemovedIn: "1.22", Replacement: "rbac.authorization.k8s.io/v1 RoleBinding", ReplaceAPIVersion: "rbac.authorization.k8s.io/v1", ReplaceKind: "RoleBinding"},

	// certificates.k8s.io/v1beta1 (removed in 1.22).
	{Group: "certificates.k8s.io", Version: "v1beta1", Kind: "CertificateSigningRequest", RemovedIn: "1.22", Replacement: "certificates.k8s.io/v1 CertificateSigningRequest", ReplaceAPIVersion: "certificates.k8s.io/v1", ReplaceKind: "CertificateSigningRequest"},

	// coordination.k8s.io/v1beta1 (removed in 1.22).
	{Group: "coordination.k8s.io", Version: "v1beta1", Kind: "Lease", RemovedIn: "1.22", Replacement: "coordination.k8s.io/v1 Lease", ReplaceAPIVersion: "coordination.k8s.io/v1", ReplaceKind: "Lease"},

	// apiregistration.k8s.io/v1beta1 (removed in 1.22).
	{Group: "apiregistration.k8s.io", Version: "v1beta1", Kind: "APIService", RemovedIn: "1.22", Replacement: "apiregistration.k8s.io/v1 APIService", ReplaceAPIVersion: "apiregistration.k8s.io/v1", ReplaceKind: "APIService"},

	// authentication.k8s.io/v1beta1 (removed in 1.22).
	{Group: "authentication.k8s.io", Version: "v1beta1", Kind: "TokenReview", RemovedIn: "1.22", Replacement: "authentication.k8s.io/v1 TokenReview", ReplaceAPIVersion: "authentication.k8s.io/v1", ReplaceKind: "TokenReview"},

	// authorization.k8s.io/v1beta1 (removed in 1.22).
	{Group: "authorization.k8s.io", Version: "v1beta1", Kind: "SubjectAccessReview", RemovedIn: "1.22", Replacement: "authorization.k8s.io/v1 SubjectAccessReview", ReplaceAPIVersion: "authorization.k8s.io/v1", ReplaceKind: "SubjectAccessReview"},

	// scheduling.k8s.io/v1beta1 (removed in 1.22).
	{Group: "scheduling.k8s.io", Version: "v1beta1", Kind: "PriorityClass", RemovedIn: "1.22", Replacement: "scheduling.k8s.io/v1 PriorityClass", ReplaceAPIVersion: "scheduling.k8s.io/v1", ReplaceKind: "PriorityClass"},

	// autoscaling/v2beta1 (removed in 1.25).
	{Group: "autoscaling", Version: "v2beta1", Kind: "HorizontalPodAutoscaler", RemovedIn: "1.25", Replacement: "autoscaling/v2 HorizontalPodAutoscaler", ReplaceAPIVersion: "autoscaling/v2", ReplaceKind: "HorizontalPodAutoscaler"},

	// batch/v1beta1 (removed in 1.25).
	{Group: "batch", Version: "v1beta1", Kind: "CronJob", RemovedIn: "1.25", Replacement: "batch/v1 CronJob", ReplaceAPIVersion: "batch/v1", ReplaceKind: "CronJob"},

	// discovery.k8s.io/v1beta1 (removed in 1.25).
	{Group: "discovery.k8s.io", Version: "v1beta1", Kind: "EndpointSlice", RemovedIn: "1.25", Replacement: "discovery.k8s.io/v1 EndpointSlice", ReplaceAPIVersion: "discovery.k8s.io/v1", ReplaceKind: "EndpointSlice"},

	// events.k8s.io/v1beta1 (removed in 1.25).
	{Group: "events.k8s.io", Version: "v1beta1", Kind: "Event", RemovedIn: "1.25", Replacement: "events.k8s.io/v1 Event", ReplaceAPIVersion: "events.k8s.io/v1", ReplaceKind: "Event"},

	// node.k8s.io/v1beta1 (removed in 1.25).
	{Group: "node.k8s.io", Version: "v1beta1", Kind: "RuntimeClass", RemovedIn: "1.25", Replacement: "node.k8s.io/v1 RuntimeClass", ReplaceAPIVersion: "node.k8s.io/v1", ReplaceKind: "RuntimeClass"},

	// policy/v1beta1 (removed in 1.25).
	{Group: "policy", Version: "v1beta1", Kind: "PodDisruptionBudget", RemovedIn: "1.25", Replacement: "policy/v1 PodDisruptionBudget", ReplaceAPIVersion: "policy/v1", ReplaceKind: "PodDisruptionBudget"},
	{Group: "policy", Version: "v1beta1", Kind: "PodSecurityPolicy", RemovedIn: "1.25", Replacement: "none (removed)"},

	// autoscaling/v2beta2 (removed in 1.26).
	{Group: "autoscaling", Version: "v2beta2", Kind: "HorizontalPodAutoscaler", RemovedIn: "1.26", Replacement: "autoscaling/v2 HorizontalPodAutoscaler", ReplaceAPIVersion: "autoscaling/v2", ReplaceKind: "HorizontalPodAutoscaler"},

	// flowcontrol.apiserver.k8s.io (removed in 1.26, 1.29 and 1.32).
	{Group: "flowcontrol.apiserver.k8s.io", Version: "v1beta1", Kind: "FlowSchema", RemovedIn: "1.26", Replacement: "flowcontrol.apiserver.k8s.io/v1 FlowSchema", ReplaceAPIVersion: "flowcontrol.apiserver.k8s.io/v1", ReplaceKind: "FlowSchema"},
	{Group: "flowcontrol.apiserver.k8s.io", Version: "v1beta1", Kind: "PriorityLevelConfiguration", RemovedIn: "1.26", Replacement: "flowcontrol.apiserver.k8s.io/v1 PriorityLevelConfiguration", ReplaceAPIVersion: "flowcontrol.apiserver.k8s.io/v1", ReplaceKind: "PriorityLevelConfiguration"},
	{Group: "flowcontrol.apiserver.k8s.io", Version: "v1beta2", Kind: "FlowSchema", RemovedIn: "1.29", Replacement: "flowcontrol.apiserver.k8s.io/v1 FlowSchema", ReplaceAPIVersion: "flowcontrol.apiserver.k8s.io/v1", ReplaceKind: "FlowSchema"},
	{Group: "flowcontrol.apiserver.k8s.io", Version: "v1beta2", Kind: "PriorityLevelConfiguration", RemovedIn: "1.29", Replacement: "flowcontrol.apiserver.k8s.io/v1 PriorityLevelConfiguration", ReplaceAPIVersion: "flowcontrol.apiserver.k8s.io/v1", ReplaceKind: "PriorityLevelConfiguration"},
	{Group: "flowcontrol.apiserver.k8s.io", Version: "v1beta3", Kind: "FlowSchema", RemovedIn: "1.32", Replacement: "flowcontrol.apiserver.k8s.io/v1 FlowSchema", ReplaceAPIVersion: "flowcontrol.apiserver.k8s.io/v1", ReplaceKind: "FlowSchema"},
	{Group: "flowcontrol.apiserver.k8s.io", Version: "v1beta3", Kind: "PriorityLevelConfiguration", RemovedIn: "1.32", Replacement: "flowcontrol.apiserver.k8s.io/v1 PriorityLevelConfiguration", ReplaceAPIVersion: "flowcontrol.apiserver.k8s.io/v1", ReplaceKind: "PriorityLevelConfiguration"},

	// storage.k8s.io/v1beta1 (CSIDriver, CSINode, StorageClass and
	// VolumeAttachment were removed in 1.22; CSIStorageCapacity in 1.27).
	{Group: "storage.k8s.io", Version: "v1beta1", Kind: "CSIDriver", RemovedIn: "1.22", Replacement: "storage.k8s.io/v1 CSIDriver", ReplaceAPIVersion: "storage.k8s.io/v1", ReplaceKind: "CSIDriver"},
	{Group: "storage.k8s.io", Version: "v1beta1", Kind: "CSINode", RemovedIn: "1.22", Replacement: "storage.k8s.io/v1 CSINode", ReplaceAPIVersion: "storage.k8s.io/v1", ReplaceKind: "CSINode"},
	{Group: "storage.k8s.io", Version: "v1beta1", Kind: "StorageClass", RemovedIn: "1.22", Replacement: "storage.k8s.io/v1 StorageClass", ReplaceAPIVersion: "storage.k8s.io/v1", ReplaceKind: "StorageClass"},
	{Group: "storage.k8s.io", Version: "v1beta1", Kind: "VolumeAttachment", RemovedIn: "1.22", Replacement: "storage.k8s.io/v1 VolumeAttachment", ReplaceAPIVersion: "storage.k8s.io/v1", ReplaceKind: "VolumeAttachment"},
	{Group: "storage.k8s.io", Version: "v1beta1", Kind: "CSIStorageCapacity", RemovedIn: "1.27", Replacement: "storage.k8s.io/v1 CSIStorageCapacity", ReplaceAPIVersion: "storage.k8s.io/v1", ReplaceKind: "CSIStorageCapacity"},
}

var kubeVersionRegexp = regexp.MustCompile(`v?(\d+)\.(\d+)`)

// parseKubeVersion extracts the major and minor components from a Kubernetes
// version string. It accepts values such as "1.25", "v1.25", "1.25.3" and
// "v1.25.3+".
func parseKubeVersion(version string) (major int, minor int, err error) {
	v := strings.TrimSpace(version)
	if v == "" {
		return 0, 0, fmt.Errorf("kubernetes version is empty")
	}

	match := kubeVersionRegexp.FindStringSubmatch(v)
	if match == nil {
		return 0, 0, fmt.Errorf("%q is not a valid kubernetes version", version)
	}

	major, err = strconv.Atoi(match[1])
	if err != nil {
		return 0, 0, err
	}

	minor, err = strconv.Atoi(match[2])
	if err != nil {
		return 0, 0, err
	}

	return major, minor, nil
}

// compareVersions compares two Kubernetes minor versions. It returns -1 when a
// is older than b, 0 when they are equal and 1 when a is newer than b. When a
// version cannot be parsed, it falls back to a lexical comparison.
func compareVersions(a, b string) int {
	aMajor, aMinor, aErr := parseKubeVersion(a)
	bMajor, bMinor, bErr := parseKubeVersion(b)
	if aErr != nil || bErr != nil {
		return strings.Compare(a, b)
	}

	switch {
	case aMajor < bMajor:
		return -1
	case aMajor > bMajor:
		return 1
	case aMinor < bMinor:
		return -1
	case aMinor > bMinor:
		return 1
	default:
		return 0
	}
}

// isRemovedInOrBefore returns true when removedIn is less than or equal to
// target, i.e. the API is gone by the time the cluster reaches target.
func isRemovedInOrBefore(removedIn, target string) bool {
	return compareVersions(removedIn, target) <= 0
}

// nextMinorVersion returns the minor version immediately after version, e.g.
// "1.27" -> "1.28".
func nextMinorVersion(version string) (string, error) {
	major, minor, err := parseKubeVersion(version)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d.%d", major, minor+1), nil
}

// latestRemovalVersion returns the newest RemovedIn version in the database.
func latestRemovalVersion() string {
	var latest string
	for _, removal := range Removals {
		if latest == "" || compareVersions(removal.RemovedIn, latest) > 0 {
			latest = removal.RemovedIn
		}
	}
	return latest
}
