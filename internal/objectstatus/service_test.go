/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectstatus

import (
	"context"
	"testing"

	linkFake "github.com/vmware-tanzu/octant/internal/link/fake"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	discoveryv1 "k8s.io/api/discovery/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/octant/internal/testutil"
	"github.com/vmware-tanzu/octant/pkg/store"
	storefake "github.com/vmware-tanzu/octant/pkg/store/fake"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

func Test_service(t *testing.T) {
	cases := []struct {
		name     string
		init     func(*testing.T, *storefake.MockStore) runtime.Object
		expected ObjectStatus
		isErr    bool
	}{
		{
			name: "in general",
			init: func(t *testing.T, o *storefake.MockStore) runtime.Object {
				slicesKey := store.Key{
					Namespace:  "default",
					APIVersion: "discovery.k8s.io/v1",
					Kind:       "EndpointSlice",
					Selector:   &labels.Set{"kubernetes.io/service-name": "stateful"},
				}
				o.EXPECT().List(gomock.Any(), gomock.Eq(slicesKey)).
					Return(testutil.ToUnstructuredList(t), false, nil)

				key := store.Key{
					Namespace:  "default",
					APIVersion: "v1",
					Kind:       "Endpoints",
					Name:       "stateful",
				}

				endpoints := testutil.LoadObjectFromFile(t, "endpoints_ok.yaml")

				o.EXPECT().Get(gomock.Any(), gomock.Eq(key)).
					Return(testutil.ToUnstructured(t, endpoints), nil)

				objectFile := "service_ok.yaml"
				return testutil.LoadObjectFromFile(t, objectFile)

			},
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusOK,
				Details:    []component.Component{component.NewText("Service is OK")},
				Properties: []component.Property{{Label: "Type", Value: component.NewText("ClusterIP")},
					{Label: "Session Affinity", Value: component.NewText("None")}},
			},
		},
		{
			name: "endpoint slices",
			init: func(t *testing.T, o *storefake.MockStore) runtime.Object {
				slicesKey := store.Key{
					Namespace:  "default",
					APIVersion: "discovery.k8s.io/v1",
					Kind:       "EndpointSlice",
					Selector:   &labels.Set{"kubernetes.io/service-name": "stateful"},
				}

				slices := []runtime.Object{
					&discoveryv1.EndpointSlice{
						TypeMeta: metav1.TypeMeta{APIVersion: "discovery.k8s.io/v1", Kind: "EndpointSlice"},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "stateful-1",
							Namespace: "default",
						},
						AddressType: discoveryv1.AddressTypeIPv4,
						Endpoints: []discoveryv1.Endpoint{
							{Addresses: []string{"10.1.85.145", "10.1.85.146"}},
							{Addresses: []string{"10.1.85.147"}},
						},
					},
					&discoveryv1.EndpointSlice{
						TypeMeta: metav1.TypeMeta{APIVersion: "discovery.k8s.io/v1", Kind: "EndpointSlice"},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "stateful-2",
							Namespace: "default",
						},
						AddressType: discoveryv1.AddressTypeIPv4,
						Endpoints: []discoveryv1.Endpoint{
							{Addresses: []string{"10.1.85.148"}},
						},
					},
				}

				o.EXPECT().List(gomock.Any(), gomock.Eq(slicesKey)).
					Return(testutil.ToUnstructuredList(t, slices...), false, nil)

				return testutil.LoadObjectFromFile(t, "service_ok.yaml")
			},
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusOK,
				Details:    []component.Component{component.NewText("Service is OK")},
				Properties: []component.Property{{Label: "Type", Value: component.NewText("ClusterIP")},
					{Label: "Session Affinity", Value: component.NewText("None")}},
			},
		},
		{
			name: "endpoint slices have no addresses",
			init: func(t *testing.T, o *storefake.MockStore) runtime.Object {
				slicesKey := store.Key{
					Namespace:  "default",
					APIVersion: "discovery.k8s.io/v1",
					Kind:       "EndpointSlice",
					Selector:   &labels.Set{"kubernetes.io/service-name": "stateful"},
				}

				slices := []runtime.Object{
					&discoveryv1.EndpointSlice{
						TypeMeta: metav1.TypeMeta{APIVersion: "discovery.k8s.io/v1", Kind: "EndpointSlice"},
						ObjectMeta: metav1.ObjectMeta{
							Name:      "stateful-1",
							Namespace: "default",
						},
						AddressType: discoveryv1.AddressTypeIPv4,
						Endpoints: []discoveryv1.Endpoint{
							{Addresses: []string{}},
						},
					},
				}

				o.EXPECT().List(gomock.Any(), gomock.Eq(slicesKey)).
					Return(testutil.ToUnstructuredList(t, slices...), false, nil)

				return testutil.LoadObjectFromFile(t, "service_ok.yaml")
			},
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusWarning,
				Details:    []component.Component{component.NewText("Service has no endpoint addresses")},
			},
		},
		{
			name: "externalName",
			init: func(t *testing.T, o *storefake.MockStore) runtime.Object {
				objectFile := "service_external.yaml"
				return testutil.LoadObjectFromFile(t, objectFile)
			},
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusOK,
				Details:    []component.Component{component.NewText("Service is OK")},
				Properties: []component.Property{{Label: "Type", Value: component.NewText("ExternalName")},
					{Label: "Session Affinity", Value: component.NewText("")}},
			},
		},
		{
			name: "no endpoint subsets",
			init: func(t *testing.T, o *storefake.MockStore) runtime.Object {
				slicesKey := store.Key{
					Namespace:  "default",
					APIVersion: "discovery.k8s.io/v1",
					Kind:       "EndpointSlice",
					Selector:   &labels.Set{"kubernetes.io/service-name": "stateful"},
				}
				o.EXPECT().List(gomock.Any(), gomock.Eq(slicesKey)).
					Return(testutil.ToUnstructuredList(t), false, nil)

				key := store.Key{
					Namespace:  "default",
					APIVersion: "v1",
					Kind:       "Endpoints",
					Name:       "stateful",
				}

				endpoints := testutil.LoadObjectFromFile(t, "endpoints_no_subsets.yaml")

				o.EXPECT().Get(gomock.Any(), gomock.Eq(key)).
					Return(testutil.ToUnstructured(t, endpoints), nil)

				objectFile := "service_ok.yaml"
				return testutil.LoadObjectFromFile(t, objectFile)

			},
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusWarning,
				Details:    []component.Component{component.NewText("Service has no endpoint addresses")},
			},
		},
		{
			name: "object is nil",
			init: func(t *testing.T, o *storefake.MockStore) runtime.Object {
				return nil
			},
			isErr: true,
		},
		{
			name: "object is not a daemon set",
			init: func(t *testing.T, o *storefake.MockStore) runtime.Object {
				return &unstructured.Unstructured{}
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			linkInterface := linkFake.NewMockInterface(controller)
			defer controller.Finish()

			o := storefake.NewMockStore(controller)

			object := tc.init(t, o)

			ctx := context.Background()
			status, err := service(ctx, object, o, linkInterface)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tc.expected, status)
		})
	}
}
