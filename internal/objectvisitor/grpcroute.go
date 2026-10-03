/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectvisitor

import (
	"context"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/vmware-tanzu/octant/internal/gvk"
	"github.com/vmware-tanzu/octant/internal/util/kubernetes"
	"github.com/vmware-tanzu/octant/pkg/store"
)

// GRPCRoute is a typed visitor for grpc routes.
type GRPCRoute struct {
	objectStore store.Store
}

var _ TypedVisitor = (*GRPCRoute)(nil)

// NewGRPCRoute creates an instance of GRPCRoute.
func NewGRPCRoute(os store.Store) *GRPCRoute {
	return &GRPCRoute{objectStore: os}
}

// Supports returns the gvk this typed visitor supports.
func (p *GRPCRoute) Supports() schema.GroupVersionKind {
	return gvk.GRPCRoute
}

// Visit visits a grpc route. It looks for backend services and parent gateways.
func (p *GRPCRoute) Visit(ctx context.Context, object *unstructured.Unstructured, handler ObjectHandler, visitor Visitor, visitDescendants bool, level int) error {
	ctx, span := trace.StartSpan(ctx, "visitGRPCRoute")
	defer span.End()

	if p.objectStore == nil {
		return errors.New("objectStore is nil")
	}

	route := &gatewayv1.GRPCRoute{}
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

	return visitReferences(ctx, p.objectStore, keys, object, handler, visitor, level, "grpc route")
}
