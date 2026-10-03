/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package clusteroverview

import (
	"context"
	"fmt"
	"sync"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/octant/internal/api"
	"github.com/vmware-tanzu/octant/internal/describer"
	"github.com/vmware-tanzu/octant/internal/gvk"
	"github.com/vmware-tanzu/octant/internal/link"
	"github.com/vmware-tanzu/octant/internal/loading"
	"github.com/vmware-tanzu/octant/internal/module"
	"github.com/vmware-tanzu/octant/internal/octant"
	"github.com/vmware-tanzu/octant/internal/printer"
	"github.com/vmware-tanzu/octant/internal/queryer"
	"github.com/vmware-tanzu/octant/pkg/config"
	"github.com/vmware-tanzu/octant/pkg/icon"
	"github.com/vmware-tanzu/octant/pkg/navigation"
	"github.com/vmware-tanzu/octant/pkg/store"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// Options are options for ClusterOverview.
type Options struct {
	DashConfig config.Dash
}

// ClusterOverview is a module for the cluster overview.
type ClusterOverview struct {
	*octant.ObjectPath
	Options

	pathMatcher *describer.PathMatcher
	watchedCRDs []*unstructured.Unstructured

	navigationCrdCache []navigation.Navigation
	navigationCrdLock  sync.Mutex

	mu sync.Mutex
}

func (co *ClusterOverview) CRDEntries(ctx context.Context, prefix, namespace string, objectStore store.Store, wantsClusterScoped bool) ([]navigation.Navigation, bool, error) {
	if co.navigationCrdCache != nil {
		return co.navigationCrdCache, false, nil
	}

	co.navigationCrdLock.Lock()
	defer co.navigationCrdLock.Unlock()

	entries, loading, err := navigation.CRDEntries(ctx, prefix, namespace, objectStore, wantsClusterScoped)
	if len(entries) > 0 {
		co.navigationCrdCache = entries
		return co.navigationCrdCache, loading, err
	}
	return []navigation.Navigation{}, false, nil
}

var _ module.Module = (*ClusterOverview)(nil)

func New(ctx context.Context, options Options) (*ClusterOverview, error) {
	pathMatcher := describer.NewPathMatcher("cluster-overview")
	for _, pf := range rootDescriber.PathFilters() {
		pathMatcher.Register(ctx, pf)
	}
	// The upgrade scanner is not a child of rootDescriber: describer.Section
	// calls Describe on every child, and we do not want to run the scan on the
	// Cluster Overview landing page. Register its path filters directly.
	for _, pf := range upgradeScannerDescriber.PathFilters() {
		pathMatcher.Register(ctx, pf)
	}

	objectPathConfig := octant.ObjectPathConfig{
		ModuleName:            "cluster-overview",
		SupportedGVKs:         supportedGVKs,
		PathLookupFunc:        gvkPath,
		ReversePathLookupFunc: gvkReversePath,
		CRDPathGenFunc:        crdPath,
	}
	objectPath, err := octant.NewObjectPath(objectPathConfig)
	if err != nil {
		return nil, errors.Wrap(err, "create module object path generator")
	}

	co := &ClusterOverview{
		ObjectPath:  objectPath,
		pathMatcher: pathMatcher,
		Options:     options,
	}

	crdWatcher := options.DashConfig.CRDWatcher()
	watchConfig := &config.CRDWatchConfig{
		Add: func(_ *describer.PathMatcher, sectionDescriber *describer.CRDSection) config.ObjectHandler {
			return func(ctx context.Context, object *unstructured.Unstructured) {
				co.mu.Lock()
				defer co.mu.Unlock()

				if object == nil {
					return
				}
				describer.AddCRD(ctx, object, pathMatcher, customResourcesDescriber, co, options.DashConfig.ObjectStore())
				co.watchedCRDs = append(co.watchedCRDs, object)
				co.navigationCrdCache = nil
			}
		}(pathMatcher, customResourcesDescriber),
		Delete: func(_ *describer.PathMatcher, csd *describer.CRDSection) config.ObjectHandler {
			return func(ctx context.Context, object *unstructured.Unstructured) {
				co.mu.Lock()
				defer co.mu.Unlock()

				if object == nil {
					return
				}
				describer.DeleteCRD(ctx, object, pathMatcher, customResourcesDescriber, co)
				var list []*unstructured.Unstructured
				for i := range co.watchedCRDs {
					if co.watchedCRDs[i].GetUID() == object.GetUID() {
						continue
					}
					list = append(list, co.watchedCRDs[i])
				}
				co.watchedCRDs = list
				co.navigationCrdCache = nil
			}
		}(pathMatcher, customResourcesDescriber),
		IsNamespaced: false,
	}

	if err := crdWatcher.AddConfig(watchConfig); err != nil {
		return nil, errors.Wrap(err, "create cluster scoped CRD watcher for cluster overview")
	}

	return co, nil
}

func (co *ClusterOverview) Name() string {
	return "cluster-overview"
}

func (co *ClusterOverview) Description() string {
	return "Cluster module is used to display all cluster related resources"
}

func (co *ClusterOverview) ClientRequestHandlers() []octant.ClientRequestHandler {
	return nil
}

func (co *ClusterOverview) Content(ctx context.Context, contentPath string, opts module.ContentOptions) (component.ContentResponse, error) {
	pf, err := co.pathMatcher.Find(contentPath)
	if err != nil {
		if err == describer.ErrPathNotFound {
			return component.EmptyContentResponse, api.NewNotFoundError(contentPath)
		}
		return component.EmptyContentResponse, err
	}

	clusterClient := co.DashConfig.ClusterClient()
	objectStore := co.DashConfig.ObjectStore()

	discoveryInterface, err := clusterClient.DiscoveryClient()
	if err != nil {
		return component.EmptyContentResponse, err
	}

	q := queryer.New(objectStore, discoveryInterface)

	p := printer.NewResource(co.DashConfig)
	if err := printer.AddHandlers(p); err != nil {
		return component.EmptyContentResponse, errors.Wrap(err, "add print handlers")
	}

	linkGenerator, err := link.NewFromDashConfig(co.DashConfig)
	if err != nil {
		return component.EmptyContentResponse, err
	}

	loaderFactory := describer.NewObjectLoaderFactory(co.DashConfig)

	options := describer.Options{
		Queryer:  q,
		Fields:   pf.Fields(contentPath),
		Printer:  p,
		LabelSet: opts.LabelSet,
		Dash:     co.DashConfig,
		Link:     linkGenerator,

		LoadObjects: loaderFactory.LoadObjects,
		LoadObject:  loaderFactory.LoadObject,
	}

	cResponse, err := pf.Describer.Describe(ctx, "", options)
	if err != nil {
		return component.EmptyContentResponse, err
	}

	return cResponse, nil
}

func (co *ClusterOverview) ContentPath() string {
	return fmt.Sprintf("%s", co.Name())
}

func (co *ClusterOverview) Navigation(ctx context.Context, _ string, root string) ([]navigation.Navigation, error) {
	navigationEntries := octant.NavigationEntries{
		Lookup: map[string]string{
			"Namespaces":                   "namespaces",
			"Custom Resources":             "custom-resources",
			"Custom Resource Definitions":  "custom-resource-definitions",
			"RBAC":                         "rbac",
			"Webhooks":                     "webhooks",
			"Workloads":                    "workloads",
			"Discovery and Load Balancing": "discovery-and-load-balancing",
			"Cluster":                      "cluster",
			"Nodes":                        "nodes",
			"Storage":                      "storage",
			"Port Forwards":                "port-forward",
			"Upgrade Scanner":              "upgrade-scanner",
		},
		EntriesFuncs: map[string]octant.EntriesFunc{
			"Cluster Overview":             nil,
			"Namespaces":                   nil,
			"Custom Resources":             co.CRDEntries,
			"Custom Resource Definitions":  nil,
			"RBAC":                         rbacEntries,
			"Webhooks":                     webhookEntries,
			"Workloads":                    workloadsEntries,
			"Discovery and Load Balancing": dlbEntries,
			"Cluster":                      clusterEntries,
			"Nodes":                        nil,
			"Storage":                      storageEntries,
			"Port Forwards":                nil,
			"Upgrade Scanner":              nil,
		},
		IconMap: map[string]string{
			"Cluster Overview":             icon.Cluster,
			"Namespaces":                   icon.Namespaces,
			"Custom Resources":             icon.CustomResources,
			"Custom Resource Definitions":  icon.CustomResourceDefinition,
			"RBAC":                         icon.RBAC,
			"Webhooks":                     icon.Webhooks,
			"Workloads":                    icon.Workloads,
			"Discovery and Load Balancing": icon.DiscoveryAndLoadBalancing,
			"Cluster":                      icon.Cluster,
			"Nodes":                        icon.Nodes,
			"Storage":                      icon.ConfigAndStorage,
			"Port Forwards":                icon.PortForwards,
			"Upgrade Scanner":              icon.Cluster,
		},
		Order: []string{
			"Cluster Overview",
			"Namespaces",
			"Custom Resources",
			"Custom Resource Definitions",
			"RBAC",
			"Webhooks",
			"Workloads",
			"Discovery and Load Balancing",
			"Cluster",
			"Nodes",
			"Storage",
			"Port Forwards",
			"Upgrade Scanner",
		},
	}

	if co.gatewayAPIAvailable() {
		navigationEntries.Lookup["Gateway API"] = "gateway-api"
		navigationEntries.EntriesFuncs["Gateway API"] = gatewayEntries
		navigationEntries.IconMap["Gateway API"] = icon.DiscoveryAndLoadBalancing
		navigationEntries.Order = append(navigationEntries.Order, "Gateway API")
	}

	objectStore := co.DashConfig.ObjectStore()

	nf := octant.NewNavigationFactory("", root, objectStore, navigationEntries)

	entries, err := nf.Generate(ctx, "Cluster", true)
	if err != nil {
		return nil, err
	}

	return entries, nil
}

// gatewayAPIAvailable returns true when the Gateway API is served by the cluster.
func (co *ClusterOverview) gatewayAPIAvailable() bool {
	if co.DashConfig == nil {
		return false
	}

	client := co.DashConfig.ClusterClient()
	if client == nil {
		return false
	}

	discoveryClient, err := client.DiscoveryClient()
	if err != nil {
		return false
	}

	resources, err := discoveryClient.ServerResourcesForGroupVersion("gateway.networking.k8s.io/v1")
	if err != nil {
		return false
	}

	for _, resource := range resources.APIResources {
		if resource.Kind == "GatewayClass" {
			return true
		}
	}

	return false
}

func (co *ClusterOverview) SetNamespace(_ string) error {
	return nil
}

func (co *ClusterOverview) Start() error {
	return nil
}

func (co *ClusterOverview) Stop() {
}

// Generators allow modules to send events to the frontend.
func (co *ClusterOverview) Generators() []octant.Generator {
	return []octant.Generator{}
}

func rbacEntries(ctx context.Context, prefix, namespace string, objectStore store.Store, _ bool) ([]navigation.Navigation, bool, error) {
	neh := navigation.EntriesHelper{}

	neh.Add("Cluster Roles", "cluster-roles",
		loading.IsObjectLoading(ctx, namespace, store.KeyFromGroupVersionKind(gvk.ClusterRole), objectStore))
	neh.Add("Cluster Role Bindings", "cluster-role-bindings",
		loading.IsObjectLoading(ctx, namespace, store.KeyFromGroupVersionKind(gvk.ClusterRoleBinding), objectStore))

	children, err := neh.Generate(prefix, namespace, "")
	if err != nil {
		return nil, false, err
	}

	return children, false, nil
}

func webhookEntries(ctx context.Context, prefix, namespace string, objectStore store.Store, _ bool) ([]navigation.Navigation, bool, error) {
	neh := navigation.EntriesHelper{}

	neh.Add("Mutating Webhooks", "mutating-webhooks",
		loading.IsObjectLoading(ctx, namespace, store.KeyFromGroupVersionKind(gvk.MutatingWebhookConfiguration), objectStore))
	neh.Add("Validating Webhooks", "validating-webhooks",
		loading.IsObjectLoading(ctx, namespace, store.KeyFromGroupVersionKind(gvk.ValidatingWebhookConfiguration), objectStore))

	children, err := neh.Generate(prefix, namespace, "")
	if err != nil {
		return nil, false, err
	}

	return children, false, nil
}

func storageEntries(ctx context.Context, prefix, namespace string, objectStore store.Store, _ bool) ([]navigation.Navigation, bool, error) {
	neh := navigation.EntriesHelper{}

	neh.Add("Persistent Volumes", "persistent-volumes",
		loading.IsObjectLoading(ctx, namespace, store.KeyFromGroupVersionKind(gvk.PersistentVolume), objectStore))
	neh.Add("Storage Classes", "storage-classes",
		loading.IsObjectLoading(ctx, namespace, store.KeyFromGroupVersionKind(gvk.StorageClass), objectStore))
	neh.Add("CSI Drivers", "csi-drivers",
		loading.IsObjectLoading(ctx, namespace, store.KeyFromGroupVersionKind(gvk.CSIDriver), objectStore))
	neh.Add("CSI Nodes", "csi-nodes",
		loading.IsObjectLoading(ctx, namespace, store.KeyFromGroupVersionKind(gvk.CSINode), objectStore))
	neh.Add("Volume Attachments", "volume-attachments",
		loading.IsObjectLoading(ctx, namespace, store.KeyFromGroupVersionKind(gvk.VolumeAttachment), objectStore))

	children, err := neh.Generate(prefix, namespace, "")
	if err != nil {
		return nil, false, err
	}

	return children, false, nil
}

func workloadsEntries(ctx context.Context, prefix, namespace string, objectStore store.Store, _ bool) ([]navigation.Navigation, bool, error) {
	neh := navigation.EntriesHelper{}

	neh.Add("Priority Classes", "priority-classes",
		loading.IsObjectLoading(ctx, namespace, store.KeyFromGroupVersionKind(gvk.PriorityClass), objectStore))
	neh.Add("Runtime Classes", "runtime-classes",
		loading.IsObjectLoading(ctx, namespace, store.KeyFromGroupVersionKind(gvk.RuntimeClass), objectStore))

	children, err := neh.Generate(prefix, namespace, "")
	if err != nil {
		return nil, false, err
	}

	return children, false, nil
}

func dlbEntries(ctx context.Context, prefix, namespace string, objectStore store.Store, _ bool) ([]navigation.Navigation, bool, error) {
	neh := navigation.EntriesHelper{}

	neh.Add("Ingress Classes", "ingress-classes",
		loading.IsObjectLoading(ctx, namespace, store.KeyFromGroupVersionKind(gvk.IngressClass), objectStore))

	children, err := neh.Generate(prefix, namespace, "")
	if err != nil {
		return nil, false, err
	}

	return children, false, nil
}

func clusterEntries(ctx context.Context, prefix, namespace string, objectStore store.Store, _ bool) ([]navigation.Navigation, bool, error) {
	neh := navigation.EntriesHelper{}

	neh.Add("Flow Schemas", "flow-schemas",
		loading.IsObjectLoading(ctx, namespace, store.KeyFromGroupVersionKind(gvk.FlowSchema), objectStore))
	neh.Add("Priority Level Configurations", "priority-level-configurations",
		loading.IsObjectLoading(ctx, namespace, store.KeyFromGroupVersionKind(gvk.PriorityLevelConfiguration), objectStore))
	neh.Add("Validating Admission Policies", "validating-admission-policies",
		loading.IsObjectLoading(ctx, namespace, store.KeyFromGroupVersionKind(gvk.ValidatingAdmissionPolicy), objectStore))

	children, err := neh.Generate(prefix, namespace, "")
	if err != nil {
		return nil, false, err
	}

	return children, false, nil
}

func gatewayEntries(ctx context.Context, prefix, namespace string, objectStore store.Store, _ bool) ([]navigation.Navigation, bool, error) {
	neh := navigation.EntriesHelper{}

	neh.Add("Gateway Classes", "gateway-classes",
		loading.IsObjectLoading(ctx, namespace, store.KeyFromGroupVersionKind(gvk.GatewayClass), objectStore))

	children, err := neh.Generate(prefix, namespace, "")
	if err != nil {
		return nil, false, err
	}

	return children, false, nil
}

func (co *ClusterOverview) SetContext(ctx context.Context, _ string) error {
	co.mu.Lock()
	defer co.mu.Unlock()

	for i := range co.watchedCRDs {
		describer.DeleteCRD(ctx, co.watchedCRDs[i], co.pathMatcher, customResourcesDescriber, co)
	}

	co.watchedCRDs = []*unstructured.Unstructured{}
	co.navigationCrdCache = nil
	return nil
}
