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

// GatewayClass is a typed visitor for gateway classes.
type GatewayClass struct {
	objectStore store.Store
}

var _ TypedVisitor = (*GatewayClass)(nil)

// NewGatewayClass creates an instance of GatewayClass.
func NewGatewayClass(os store.Store) *GatewayClass {
	return &GatewayClass{objectStore: os}
}

// Supports returns the gvk this typed visitor supports.
func (p *GatewayClass) Supports() schema.GroupVersionKind {
	return gvk.GatewayClass
}

// Visit visits a gateway class. It looks for gateways referencing the class.
func (p *GatewayClass) Visit(ctx context.Context, object *unstructured.Unstructured, handler ObjectHandler, visitor Visitor, visitDescendants bool, level int) error {
	ctx, span := trace.StartSpan(ctx, "visitGatewayClass")
	defer span.End()

	if p.objectStore == nil {
		return errors.New("objectStore is nil")
	}

	gatewayClass := &gatewayv1.GatewayClass{}
	if err := kubernetes.FromUnstructured(object, gatewayClass); err != nil {
		return err
	}
	level = handler.SetLevel(gatewayClass.Kind, level)

	list, _, err := p.objectStore.List(ctx, store.KeyFromGroupVersionKind(gvk.Gateway))
	if err != nil {
		if kerrors.IsNotFound(err) {
			return nil
		}
		return err
	}
	if list == nil {
		return nil
	}

	var g errgroup.Group

	for i := range list.Items {
		item := list.Items[i]
		name, _, err := unstructured.NestedString(item.Object, "spec", "gatewayClassName")
		if err != nil || name != gatewayClass.Name {
			continue
		}

		gateway := item.DeepCopy()
		g.Go(func() error {
			if err := visitor.Visit(ctx, gateway, handler, true, level); err != nil {
				return errors.Wrapf(err, "gateway class %s visit gateway %s",
					kubernetes.PrintObject(gatewayClass), kubernetes.PrintObject(gateway))
			}

			return handler.AddEdge(ctx, object, gateway, level)
		})
	}

	return g.Wait()
}
