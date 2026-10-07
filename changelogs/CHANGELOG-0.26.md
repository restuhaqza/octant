## v0.26.0

#### 2026-10-03

### Download

- https://github.com/restuhaqza/octant/releases/tag/v0.26.0

### All Changes

- Bumped Kubernetes client libraries (`k8s.io/api`, `k8s.io/apimachinery`, `k8s.io/client-go`, `k8s.io/apiextensions-apiserver`, `k8s.io/kube-aggregator`, `k8s.io/metrics`) to `v0.34.1` and the Go toolchain to 1.24.
- Migrated CronJob views to the `batch/v1` API group and removed support for the removed `extensions/v1beta1` Deployment and ReplicaSet groups.
- Added `autoscaling/v2` HorizontalPodAutoscaler support (list, detail, status, metrics, conditions, and configuration including native scale behavior) and made it the canonical HPA version; `autoscaling/v1` continues to work.
- Added Gateway API support (`gateway.networking.k8s.io/v1`) for Gateways, HTTPRoutes, GRPCRoutes, and GatewayClasses, including navigation, a cluster overview entry, list and detail views, object status, and resource viewer edges. These are only shown when the Gateway API is served by the cluster.
- Prefer `discovery.k8s.io/v1` EndpointSlices when rendering Service endpoints and computing Service status, falling back to the legacy core `v1` Endpoints object when no EndpointSlices exist for the Service. Only ready endpoints are counted and listed.
- Added native list and detail views, navigation entries, content paths, and object status for previously unsupported stable Kubernetes resources: PodDisruptionBudget, ResourceQuota, LimitRange, Lease, CSIStorageCapacity, PriorityClass, RuntimeClass, IngressClass, CSIDriver, CSINode, VolumeAttachment, FlowSchema, PriorityLevelConfiguration, and ValidatingAdmissionPolicy; namespace detail views now also show a Pod Security summary.
- Added a read-only Upgrade & Deprecation Scanner view that finds live objects using Kubernetes API versions removed or deprecated in a selectable target version, with object links to the replacement API, per-group counts, and API-group filtering.
