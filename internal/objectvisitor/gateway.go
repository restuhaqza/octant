/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectvisitor

import (
	"context"

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

// Gateway is a typed visitor for gateways.
type Gateway struct {
	objectStore store.Store
}

var _ TypedVisitor = (*Gateway)(nil)

// NewGateway creates an instance of Gateway.
func NewGateway(os store.Store) *Gateway {
	return &Gateway{objectStore: os}
}

// Supports returns the gvk this typed visitor supports.
func (p *Gateway) Supports() schema.GroupVersionKind {
	return gvk.Gateway
}

// Visit visits a gateway. It looks for the referenced gateway class.
func (p *Gateway) Visit(ctx context.Context, object *unstructured.Unstructured, handler ObjectHandler, visitor Visitor, visitDescendants bool, level int) error {
	ctx, span := trace.StartSpan(ctx, "visitGateway")
	defer span.End()

	if p.objectStore == nil {
		return errors.New("objectStore is nil")
	}

	gateway := &gatewayv1.Gateway{}
	if err := kubernetes.FromUnstructured(object, gateway); err != nil {
		return err
	}
	level = handler.SetLevel(gateway.Kind, level)

	className := string(gateway.Spec.GatewayClassName)
	if className == "" {
		return nil
	}

	var g errgroup.Group

	g.Go(func() error {
		key := store.KeyFromGroupVersionKind(gvk.GatewayClass)
		key.Name = className
		gatewayClass, err := p.objectStore.Get(ctx, key)
		if err != nil {
			if kerrors.IsNotFound(err) {
				return nil
			}
			return err
		}
		if gatewayClass == nil {
			return nil
		}

		if err := visitor.Visit(ctx, gatewayClass, handler, true, level); err != nil {
			return errors.Wrapf(err, "gateway %s visit gateway class %s",
				kubernetes.PrintObject(gateway), kubernetes.PrintObject(gatewayClass))
		}

		return handler.AddEdge(ctx, object, gatewayClass, level)
	})

	return g.Wait()
}

// backendServiceKey builds an object store key for a Service backend reference.
// The second return value is false when the reference is not a core Service.
func backendServiceKey(defaultNamespace string, group *gatewayv1.Group, kind *gatewayv1.Kind, namespace *gatewayv1.Namespace, name string) (store.Key, bool) {
	objectKind := "Service"
	if kind != nil && string(*kind) != "" {
		objectKind = string(*kind)
	}

	if objectKind != "Service" {
		return store.Key{}, false
	}

	if group != nil && string(*group) != "" {
		return store.Key{}, false
	}

	if name == "" {
		return store.Key{}, false
	}

	ns := defaultNamespace
	if namespace != nil && string(*namespace) != "" {
		ns = string(*namespace)
	}

	return store.Key{APIVersion: "v1", Kind: "Service", Namespace: ns, Name: name}, true
}

// parentGatewayKey builds an object store key for a parent Gateway reference.
// The second return value is false when the reference is not a Gateway.
func parentGatewayKey(defaultNamespace string, group *gatewayv1.Group, kind *gatewayv1.Kind, namespace *gatewayv1.Namespace, name string) (store.Key, bool) {
	objectKind := "Gateway"
	if kind != nil && string(*kind) != "" {
		objectKind = string(*kind)
	}

	if objectKind != "Gateway" {
		return store.Key{}, false
	}

	if group != nil && string(*group) != "" && string(*group) != gvk.Gateway.Group {
		return store.Key{}, false
	}

	if name == "" {
		return store.Key{}, false
	}

	ns := defaultNamespace
	if namespace != nil && string(*namespace) != "" {
		ns = string(*namespace)
	}

	key := store.KeyFromGroupVersionKind(gvk.Gateway)
	key.Namespace = ns
	key.Name = name

	return key, true
}
