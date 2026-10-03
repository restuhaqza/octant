/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package upgradescanner

import (
	"context"
	"sort"
	"strings"

	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"

	"github.com/vmware-tanzu/octant/pkg/store"
)

// maxObjectsPerGVK bounds how many live objects are reported for a single
// group/version/kind so a single noisy API cannot dominate the view.
const maxObjectsPerGVK = 200

// Finding is a live object that uses an API version that is deprecated or
// removed in the target Kubernetes version.
type Finding struct {
	Group       string
	Version     string
	Kind        string
	RemovedIn   string
	Replacement string
	Namespace   string
	Name        string

	// ReplaceAPIVersion and ReplaceKind describe the replacement API used to
	// generate an object link. They are empty when there is no replacement.
	ReplaceAPIVersion string
	ReplaceKind       string

	Removed bool // true if RemovedIn <= target
}

// GroupVersion returns the apiVersion string for the finding, e.g.
// "batch/v1beta1" or "v1" for the core group.
func (f Finding) GroupVersion() string {
	if f.Group == "" {
		return f.Version
	}
	return f.Group + "/" + f.Version
}

type removalKey struct {
	group   string
	version string
	kind    string
}

var removalLookup = buildRemovalLookup(Removals)

func buildRemovalLookup(removals []APIRemoval) map[removalKey]APIRemoval {
	lookup := make(map[removalKey]APIRemoval, len(removals))
	for _, removal := range removals {
		lookup[removalKey{group: removal.Group, version: removal.Version, kind: removal.Kind}] = removal
	}
	return lookup
}

// Scan enumerates the API resources served by the cluster and reports live
// objects that use an API version present in the deprecation database. It only
// reads from the object store and never writes to the cluster.
//
// namespace may be empty, in which case namespaced resources are listed across
// all namespaces (cluster-wide). When namespace is set, namespaced resources
// are scoped to that namespace.
func Scan(ctx context.Context, discoveryClient discovery.DiscoveryInterface, objectStore store.Store, namespace, target string) ([]Finding, error) {
	if discoveryClient == nil {
		return nil, errors.New("discovery client is nil")
	}
	if objectStore == nil {
		return nil, errors.New("object store is nil")
	}

	_, resourceLists, err := discoveryClient.ServerGroupsAndResources()
	// ServerGroupsAndResources can return partial results together with an
	// error when one API group is unavailable. Only fail when we have nothing
	// to work with.
	if err != nil && len(resourceLists) == 0 {
		return nil, err
	}

	var findings []Finding
	for _, resourceList := range resourceLists {
		if resourceList == nil {
			continue
		}

		groupVersion, err := schema.ParseGroupVersion(resourceList.GroupVersion)
		if err != nil {
			continue
		}

		for _, resource := range resourceList.APIResources {
			if resource.Name == "" || resource.Kind == "" {
				continue
			}
			// Skip subresources such as "cronjobs/status".
			if strings.Contains(resource.Name, "/") {
				continue
			}
			// Skip list pseudo-kinds.
			if strings.HasSuffix(resource.Kind, "List") {
				continue
			}

			removal, ok := removalLookup[removalKey{
				group:   groupVersion.Group,
				version: groupVersion.Version,
				kind:    resource.Kind,
			}]
			if !ok {
				continue
			}

			key := store.Key{APIVersion: resourceList.GroupVersion, Kind: resource.Kind}
			if namespace != "" && resource.Namespaced {
				key.Namespace = namespace
			}

			list, _, err := objectStore.List(ctx, key)
			// Ignore list errors (for example missing informers) so the view
			// still renders the remaining findings.
			if err != nil || list == nil {
				continue
			}

			count := 0
			for i := range list.Items {
				if count >= maxObjectsPerGVK {
					break
				}

				object := list.Items[i]
				findings = append(findings, Finding{
					Group:             groupVersion.Group,
					Version:           groupVersion.Version,
					Kind:              resource.Kind,
					RemovedIn:         removal.RemovedIn,
					Replacement:       removal.Replacement,
					Namespace:         object.GetNamespace(),
					Name:              object.GetName(),
					ReplaceAPIVersion: removal.ReplaceAPIVersion,
					ReplaceKind:       removal.ReplaceKind,
					Removed:           isRemovedInOrBefore(removal.RemovedIn, target),
				})
				count++
			}
		}
	}

	sortFindings(findings)

	return findings, nil
}

func sortFindings(findings []Finding) {
	sort.Slice(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.Group != b.Group {
			return a.Group < b.Group
		}
		if a.Version != b.Version {
			return a.Version < b.Version
		}
		if a.Kind != b.Kind {
			return a.Kind < b.Kind
		}
		if a.Namespace != b.Namespace {
			return a.Namespace < b.Namespace
		}
		return a.Name < b.Name
	})
}
