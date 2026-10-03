/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectstatus

import (
	"context"
	"sort"
	"strings"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/vmware-tanzu/octant/internal/link"
	"github.com/vmware-tanzu/octant/pkg/store"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

func resourceQuota(_ context.Context, object runtime.Object, _ store.Store, _ link.Interface) (ObjectStatus, error) {
	if object == nil {
		return ObjectStatus{}, errors.New("resource quota is nil")
	}

	quota := &corev1.ResourceQuota{}

	if err := scheme.Scheme.Convert(object, quota, 0); err != nil {
		return ObjectStatus{}, errors.Wrap(err, "convert object to v1 ResourceQuota")
	}

	var exceeded []string
	for name, hard := range quota.Status.Hard {
		used, ok := quota.Status.Used[name]
		if !ok {
			continue
		}

		if used.Cmp(hard) > 0 {
			exceeded = append(exceeded, name.String())
		}
	}

	if len(exceeded) > 0 {
		sort.Strings(exceeded)
		return ObjectStatus{
			NodeStatus: component.NodeStatusError,
			Details: []component.Component{
				component.NewTextf("Exceeded quota: %s", strings.Join(exceeded, ", ")),
			},
		}, nil
	}

	return ObjectStatus{
		NodeStatus: component.NodeStatusOK,
		Details:    []component.Component{component.NewText("v1 ResourceQuota is OK")},
	}, nil
}
