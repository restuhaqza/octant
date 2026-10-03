/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	schedulingv1 "k8s.io/api/scheduling/v1"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// PriorityClassListHandler is a printFunc that prints priority classes.
func PriorityClassListHandler(ctx context.Context, list *schedulingv1.PriorityClassList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("priority class list is nil")
	}

	cols := component.NewTableCols("Name", "Value", "Global Default", "Age")
	ot := NewObjectTable("Priority Classes", "We couldn't find any priority classes!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())
	for _, priorityClass := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&priorityClass, priorityClass.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Value"] = component.NewText(fmt.Sprintf("%d", priorityClass.Value))
		row["Global Default"] = component.NewText(fmt.Sprintf("%v", priorityClass.GlobalDefault))
		row["Age"] = component.NewTimestamp(priorityClass.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &priorityClass, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// PriorityClassHandler is a printFunc that prints a single priority class.
func PriorityClassHandler(ctx context.Context, priorityClass *schedulingv1.PriorityClass, options Options) (component.Component, error) {
	o := NewObject(priorityClass)
	o.EnableEvents()

	config, err := priorityClassConfig(priorityClass)
	if err != nil {
		return nil, err
	}
	o.RegisterConfig(config)

	return o.ToComponent(ctx, options)
}

func priorityClassConfig(priorityClass *schedulingv1.PriorityClass) (*component.Summary, error) {
	if priorityClass == nil {
		return nil, errors.New("priority class is nil")
	}

	sections := component.SummarySections{}
	sections.AddText("Value", fmt.Sprintf("%d", priorityClass.Value))
	sections.AddText("Global Default", fmt.Sprintf("%v", priorityClass.GlobalDefault))
	sections.AddText("Description", priorityClass.Description)
	if priorityClass.PreemptionPolicy != nil {
		sections.AddText("Preemption Policy", string(*priorityClass.PreemptionPolicy))
	}

	return component.NewSummary("Configuration", sections...), nil
}
