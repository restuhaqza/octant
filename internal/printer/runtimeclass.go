/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	nodev1 "k8s.io/api/node/v1"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// RuntimeClassListHandler is a printFunc that prints runtime classes.
func RuntimeClassListHandler(ctx context.Context, list *nodev1.RuntimeClassList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("runtime class list is nil")
	}

	cols := component.NewTableCols("Name", "Handler", "Age")
	ot := NewObjectTable("Runtime Classes", "We couldn't find any runtime classes!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())
	for _, runtimeClass := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&runtimeClass, runtimeClass.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Handler"] = component.NewText(runtimeClass.Handler)
		row["Age"] = component.NewTimestamp(runtimeClass.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &runtimeClass, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// RuntimeClassHandler is a printFunc that prints a single runtime class.
func RuntimeClassHandler(ctx context.Context, runtimeClass *nodev1.RuntimeClass, options Options) (component.Component, error) {
	o := NewObject(runtimeClass)
	o.EnableEvents()

	config, err := runtimeClassConfig(runtimeClass)
	if err != nil {
		return nil, err
	}
	o.RegisterConfig(config)

	return o.ToComponent(ctx, options)
}

func runtimeClassConfig(runtimeClass *nodev1.RuntimeClass) (*component.Summary, error) {
	if runtimeClass == nil {
		return nil, errors.New("runtime class is nil")
	}

	sections := component.SummarySections{}
	sections.AddText("Handler", runtimeClass.Handler)
	if runtimeClass.Overhead != nil {
		sections.AddText("Overhead CPU", runtimeClass.Overhead.PodFixed.Cpu().String())
		sections.AddText("Overhead Memory", runtimeClass.Overhead.PodFixed.Memory().String())
	}
	if runtimeClass.Scheduling != nil {
		sections.Add("Node Selector", printSelectorMap(runtimeClass.Scheduling.NodeSelector))
	}

	return component.NewSummary("Configuration", sections...), nil
}
