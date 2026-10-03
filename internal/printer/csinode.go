/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	storagev1 "k8s.io/api/storage/v1"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// CSINodeListHandler is a printFunc that prints CSI nodes.
func CSINodeListHandler(ctx context.Context, list *storagev1.CSINodeList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("csi node list is nil")
	}

	cols := component.NewTableCols("Name", "Drivers", "Age")
	ot := NewObjectTable("CSI Nodes", "We couldn't find any csi nodes!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())
	for _, csiNode := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&csiNode, csiNode.Name)
		if err != nil {
			return nil, err
		}

		names := make([]string, len(csiNode.Spec.Drivers))
		for i := range csiNode.Spec.Drivers {
			names[i] = csiNode.Spec.Drivers[i].Name
		}

		row["Name"] = nameLink
		row["Drivers"] = component.NewText(strings.Join(names, ", "))
		row["Age"] = component.NewTimestamp(csiNode.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &csiNode, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// CSINodeHandler is a printFunc that prints a single CSI node.
func CSINodeHandler(ctx context.Context, csiNode *storagev1.CSINode, options Options) (component.Component, error) {
	o := NewObject(csiNode)
	o.EnableEvents()

	o.RegisterItems(ItemDescriptor{
		Width: component.WidthFull,
		Func: func() (component.Component, error) {
			return csiNodeDrivers(csiNode)
		},
	})

	return o.ToComponent(ctx, options)
}

func csiNodeDrivers(csiNode *storagev1.CSINode) (*component.Table, error) {
	if csiNode == nil {
		return nil, errors.New("csi node is nil")
	}

	cols := component.NewTableCols("Name", "Node ID", "Topology Keys")
	table := component.NewTable("Drivers", "There are no drivers!", cols)

	for i := range csiNode.Spec.Drivers {
		driver := csiNode.Spec.Drivers[i]
		table.Add(component.TableRow{
			"Name":          component.NewText(driver.Name),
			"Node ID":       component.NewText(driver.NodeID),
			"Topology Keys": component.NewText(strings.Join(driver.TopologyKeys, ", ")),
		})
	}

	table.Sort("Name")

	return table, nil
}
