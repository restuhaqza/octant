/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectvisitor

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"golang.org/x/sync/errgroup"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/vmware-tanzu/octant/internal/gvk"
	"github.com/vmware-tanzu/octant/internal/util/kubernetes"
	"github.com/vmware-tanzu/octant/pkg/store"
)

// HTTPRoute is a typed visitor for http routes.
type HTTPRoute struct {
	objectStore store.Store
}

var _ TypedVisitor = (*HTTPRoute)(nil)

// NewHTTPRoute creates an instance of HTTPRoute.
func NewHTTPRoute(os store.Store) *HTTPRoute {
	return &HTTPRoute{objectStore: os}
}

// Supports returns the gvk this typed visitor supports.
func (p *HTTPRoute) Supports() schema.GroupVersionKind {
	return gvk.HTTPRoute
}

// Visit visits an http route. It looks for backend services and parent gateways.
func (p *HTTPRoute) Visit(ctx context.Context, object *unstructured.Unstructured, handler ObjectHandler, visitor Visitor, visitDescendants bool, level int) error {
	ctx, span := trace.StartSpan(ctx, "visitHTTPRoute")
	defer span.End()

	if p.objectStore == nil {
		return errors.New("objectStore is nil")
	}

	route := &gatewayv1.HTTPRoute{}
	if err := kubernetes.FromUnstructured(object, route); err != nil {
		return err
	}
	level = handler.SetLevel(route.Kind, level)

	keys := map[string]store.Key{}

	for _, rule := range route.Spec.Rules {
		for _, ref := range rule.BackendRefs {
			if key, ok := backendServiceKey(route.Namespace, ref.Group, ref.Kind, ref.Namespace, string(ref.Name)); ok {
				keys[storeKeyString(key)] = key
			}
		}
	}

	for _, ref := range route.Spec.ParentRefs {
		if key, ok := parentGatewayKey(route.Namespace, ref.Group, ref.Kind, ref.Namespace, string(ref.Name)); ok {
			keys[storeKeyString(key)] = key
		}
	}

	return visitReferences(ctx, p.objectStore, keys, object, handler, visitor, level, "http route")
}

// storeKeyString returns a stable string representation of a store key.
func storeKeyString(key store.Key) string {
	return fmt.Sprintf("%s/%s/%s/%s", key.APIVersion, key.Kind, key.Namespace, key.Name)
}

// visitReferences fetches each key from the object store and records an edge.
func visitReferences(ctx context.Context, objectStore store.Store, keys map[string]store.Key, source *unstructured.Unstructured, handler ObjectHandler, visitor Visitor, level int, sourceKind string) error {
	var g errgroup.Group

	for _, key := range keys {
		k := key
		g.Go(func() error {
			target, err := objectStore.Get(ctx, k)
			if err != nil {
				if kerrors.IsNotFound(err) {
					return nil
				}
				return err
			}
			if target == nil {
				return nil
			}

			if err := visitor.Visit(ctx, target, handler, true, level); err != nil {
				return errors.Wrapf(err, "%s %s visit %s %s",
					sourceKind, kubernetes.PrintObject(source), k.Kind, kubernetes.PrintObject(target))
			}

			return handler.AddEdge(ctx, source, target, level)
		})
	}

	return g.Wait()
}
