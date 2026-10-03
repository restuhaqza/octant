/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectstatus

import (
	"context"

	"github.com/vmware-tanzu/octant/internal/link"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/vmware-tanzu/octant/pkg/store"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

func service(ctx context.Context, object runtime.Object, o store.Store, _ link.Interface) (ObjectStatus, error) {
	if object == nil {
		return ObjectStatus{}, errors.Errorf("service is nil")
	}

	service := &corev1.Service{}

	if err := scheme.Scheme.Convert(object, service, 0); err != nil {
		return ObjectStatus{}, errors.Wrap(err, "convert object to service")
	}

	if service.Spec.ExternalName == "" {
		endpointSlices, err := listEndpointSlicesForService(ctx, service, o)
		if err != nil {
			return ObjectStatus{}, errors.Wrapf(err, "list endpoint slices for service %s", service.Name)
		}

		if len(endpointSlices) > 0 {
			addressCount := 0

			for _, endpointSlice := range endpointSlices {
				for _, endpoint := range endpointSlice.Endpoints {
					addressCount += len(endpoint.Addresses)
				}
			}

			if addressCount == 0 {
				return ObjectStatus{
					NodeStatus: component.NodeStatusWarning,
					Details:    []component.Component{component.NewText("Service has no endpoint addresses")},
				}, nil
			}
		} else {
			key := store.Key{
				Namespace:  service.Namespace,
				APIVersion: "v1",
				Kind:       "Endpoints",
				Name:       service.Name,
			}

			endpoints := &corev1.Endpoints{}

			found, err := store.GetAs(ctx, o, key, endpoints)
			if err != nil {
				return ObjectStatus{}, errors.Wrapf(err, "get endpoints for service %s", service.Name)
			}

			if !found {
				return ObjectStatus{
					NodeStatus: component.NodeStatusWarning,
					Details:    []component.Component{component.NewText("Service has no endpoints")},
				}, nil
			}

			addressCount := 0

			for _, subset := range endpoints.Subsets {
				addressCount += len(subset.Addresses)
			}

			if addressCount == 0 {
				return ObjectStatus{
					NodeStatus: component.NodeStatusWarning,
					Details:    []component.Component{component.NewText("Service has no endpoint addresses")},
				}, nil
			}
		}
	}
	properties := []component.Property{{Label: "Type", Value: component.NewText(string(service.Spec.Type))},
		{Label: "Session Affinity", Value: component.NewText(string(service.Spec.SessionAffinity))}}

	return ObjectStatus{
		NodeStatus: component.NodeStatusOK,
		Details:    []component.Component{component.NewText("Service is OK")},
		Properties: properties,
	}, nil
}

// listEndpointSlicesForService lists the discovery.k8s.io/v1 EndpointSlices
// belonging to a service. EndpointSlices are labelled with the owning service
// name.
func listEndpointSlicesForService(ctx context.Context, service *corev1.Service, o store.Store) ([]*discoveryv1.EndpointSlice, error) {
	if service == nil {
		return nil, errors.New("service is nil")
	}

	serviceNameLabel := labels.Set{"kubernetes.io/service-name": service.Name}
	key := store.Key{
		Namespace:  service.Namespace,
		APIVersion: "discovery.k8s.io/v1",
		Kind:       "EndpointSlice",
		Selector:   &serviceNameLabel,
	}

	list, _, err := o.List(ctx, key)
	if err != nil {
		return nil, errors.Wrapf(err, "list endpoint slices for service %s", service.Name)
	}

	var endpointSlices []*discoveryv1.EndpointSlice
	for i := range list.Items {
		endpointSlice := &discoveryv1.EndpointSlice{}
		if err := scheme.Scheme.Convert(&list.Items[i], endpointSlice, 0); err != nil {
			return nil, errors.Wrap(err, "convert unstructured object to endpoint slice")
		}
		endpointSlices = append(endpointSlices, endpointSlice)
	}

	return endpointSlices, nil
}
