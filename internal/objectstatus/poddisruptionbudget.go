/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectstatus

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	policyv1 "k8s.io/api/policy/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/vmware-tanzu/octant/internal/link"
	"github.com/vmware-tanzu/octant/pkg/store"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

func podDisruptionBudget(_ context.Context, object runtime.Object, _ store.Store, _ link.Interface) (ObjectStatus, error) {
	if object == nil {
		return ObjectStatus{}, errors.New("pod disruption budget is nil")
	}

	pdb := &policyv1.PodDisruptionBudget{}

	if err := scheme.Scheme.Convert(object, pdb, 0); err != nil {
		return ObjectStatus{}, errors.Wrap(err, "convert object to v1 PodDisruptionBudget")
	}

	if pdb.Status.CurrentHealthy < pdb.Status.DesiredHealthy {
		return ObjectStatus{
			NodeStatus: component.NodeStatusWarning,
			Details: []component.Component{
				component.NewText(fmt.Sprintf("%d current healthy, %d desired healthy",
					pdb.Status.CurrentHealthy, pdb.Status.DesiredHealthy)),
			},
		}, nil
	}

	return ObjectStatus{
		NodeStatus: component.NodeStatusOK,
		Details: []component.Component{
			component.NewText(fmt.Sprintf("%d disruptions allowed", pdb.Status.DisruptionsAllowed)),
		},
	}, nil
}
