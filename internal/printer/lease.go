/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	coordinationv1 "k8s.io/api/coordination/v1"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// LeaseListHandler is a printFunc that prints leases.
func LeaseListHandler(ctx context.Context, list *coordinationv1.LeaseList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("lease list is nil")
	}

	cols := component.NewTableCols("Name", "Holder", "Age")
	ot := NewObjectTable("Leases", "We couldn't find any leases!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())
	for _, lease := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&lease, lease.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Holder"] = component.NewText(formatStringPtr(lease.Spec.HolderIdentity))
		row["Age"] = component.NewTimestamp(lease.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &lease, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// LeaseHandler is a printFunc that prints a single lease.
func LeaseHandler(ctx context.Context, lease *coordinationv1.Lease, options Options) (component.Component, error) {
	o := NewObject(lease)
	o.EnableEvents()

	config, err := leaseConfig(lease)
	if err != nil {
		return nil, err
	}
	o.RegisterConfig(config)

	return o.ToComponent(ctx, options)
}

func leaseConfig(lease *coordinationv1.Lease) (*component.Summary, error) {
	if lease == nil {
		return nil, errors.New("lease is nil")
	}

	sections := component.SummarySections{}
	sections.AddText("Holder", formatStringPtr(lease.Spec.HolderIdentity))
	sections.AddText("Lease Duration (seconds)", formatInt32Ptr(lease.Spec.LeaseDurationSeconds))
	if lease.Spec.AcquireTime != nil {
		sections.Add("Acquire Time", component.NewTimestamp(lease.Spec.AcquireTime.Time))
	}
	if lease.Spec.RenewTime != nil {
		sections.Add("Renew Time", component.NewTimestamp(lease.Spec.RenewTime.Time))
	}
	sections.AddText("Lease Transitions", formatInt32Ptr(lease.Spec.LeaseTransitions))

	return component.NewSummary("Configuration", sections...), nil
}
