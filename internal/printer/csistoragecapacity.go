/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	storagev1 "k8s.io/api/storage/v1"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// CSIStorageCapacityListHandler is a printFunc that prints CSI storage capacities.
func CSIStorageCapacityListHandler(ctx context.Context, list *storagev1.CSIStorageCapacityList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("csi storage capacity list is nil")
	}

	cols := component.NewTableCols("Name", "Storage Class", "Capacity", "Age")
	ot := NewObjectTable("CSI Storage Capacities", "We couldn't find any csi storage capacities!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())
	for _, capacity := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&capacity, capacity.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Storage Class"] = component.NewText(capacity.StorageClassName)
		row["Capacity"] = component.NewText(formatQuantityPtr(capacity.Capacity))
		row["Age"] = component.NewTimestamp(capacity.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &capacity, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// CSIStorageCapacityHandler is a printFunc that prints a single CSI storage capacity.
func CSIStorageCapacityHandler(ctx context.Context, capacity *storagev1.CSIStorageCapacity, options Options) (component.Component, error) {
	o := NewObject(capacity)
	o.EnableEvents()

	config, err := csiStorageCapacityConfig(capacity)
	if err != nil {
		return nil, err
	}
	o.RegisterConfig(config)

	return o.ToComponent(ctx, options)
}

func csiStorageCapacityConfig(capacity *storagev1.CSIStorageCapacity) (*component.Summary, error) {
	if capacity == nil {
		return nil, errors.New("csi storage capacity is nil")
	}

	sections := component.SummarySections{}
	sections.AddText("Storage Class", capacity.StorageClassName)
	sections.AddText("Capacity", formatQuantityPtr(capacity.Capacity))
	sections.AddText("Maximum Volume Size", formatQuantityPtr(capacity.MaximumVolumeSize))
	sections.Add("Node Topology", printSelector(capacity.NodeTopology))

	return component.NewSummary("Configuration", sections...), nil
}
