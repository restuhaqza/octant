/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectvisitor_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/kubernetes/scheme"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/vmware-tanzu/octant/internal/objectvisitor"
	"github.com/vmware-tanzu/octant/internal/objectvisitor/fake"
	"github.com/vmware-tanzu/octant/internal/testutil"
	"github.com/vmware-tanzu/octant/pkg/store"
	objectStoreFake "github.com/vmware-tanzu/octant/pkg/store/fake"
)

func TestHTTPRoute_Visit(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	service := testutil.CreateService("service")

	object := testutil.CreateHTTPRoute("http-route")
	object.Spec.Rules = []gatewayv1.HTTPRouteRule{
		{
			BackendRefs: []gatewayv1.HTTPBackendRef{
				{
					BackendRef: gatewayv1.BackendRef{
						BackendObjectReference: gatewayv1.BackendObjectReference{
							Name: "service",
						},
					},
				},
			},
		},
	}
	u := testutil.ToUnstructured(t, object)

	handler := fake.NewMockObjectHandler(controller)
	handler.EXPECT().SetLevel(gomock.Any(), 1).Return(2)
	handler.EXPECT().
		AddEdge(gomock.Any(), u, testutil.ToUnstructured(t, service), gomock.Any()).
		Return(nil)

	var visited []unstructured.Unstructured
	visitor := fake.NewMockVisitor(controller)
	visitor.EXPECT().
		Visit(gomock.Any(), gomock.Any(), handler, true, gomock.Any()).
		DoAndReturn(func(ctx context.Context, object *unstructured.Unstructured, handler objectvisitor.ObjectHandler, _ bool, _ int) error {
			visited = append(visited, *object)
			return nil
		})

	objectStore := objectStoreFake.NewMockStore(controller)

	key := store.Key{
		APIVersion: "v1",
		Kind:       "Service",
		Namespace:  service.Namespace,
		Name:       service.Name,
	}
	objectStore.EXPECT().
		Get(gomock.Any(), key).
		Return(testutil.ToUnstructured(t, service), nil)

	httpRoute := objectvisitor.NewHTTPRoute(objectStore)

	ctx := context.Background()
	err := httpRoute.Visit(ctx, u, handler, visitor, true, 1)

	sortObjectsByName(t, visited)

	expected := testutil.ToUnstructuredList(t, service)
	assert.Equal(t, expected.Items, visited)
	assert.NoError(t, err)
}

func init() {
	_ = gatewayv1.Install(scheme.Scheme)
}
