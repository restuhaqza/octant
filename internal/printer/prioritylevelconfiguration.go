/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	flowcontrolv1 "k8s.io/api/flowcontrol/v1"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// PriorityLevelConfigurationListHandler is a printFunc that prints priority level configurations.
func PriorityLevelConfigurationListHandler(ctx context.Context, list *flowcontrolv1.PriorityLevelConfigurationList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("priority level configuration list is nil")
	}

	cols := component.NewTableCols("Name", "Type", "Age")
	ot := NewObjectTable("Priority Level Configurations", "We couldn't find any priority level configurations!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())
	for _, config := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&config, config.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Type"] = component.NewText(string(config.Spec.Type))
		row["Age"] = component.NewTimestamp(config.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &config, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// PriorityLevelConfigurationHandler is a printFunc that prints a single priority level configuration.
func PriorityLevelConfigurationHandler(ctx context.Context, config *flowcontrolv1.PriorityLevelConfiguration, options Options) (component.Component, error) {
	o := NewObject(config)
	o.EnableEvents()

	summary, err := priorityLevelConfigurationConfig(config)
	if err != nil {
		return nil, err
	}
	o.RegisterConfig(summary)

	return o.ToComponent(ctx, options)
}

func priorityLevelConfigurationConfig(config *flowcontrolv1.PriorityLevelConfiguration) (*component.Summary, error) {
	if config == nil {
		return nil, errors.New("priority level configuration is nil")
	}

	sections := component.SummarySections{}
	sections.AddText("Type", string(config.Spec.Type))
	if limited := config.Spec.Limited; limited != nil {
		sections.AddText("Nominal Concurrency Shares", formatInt32Ptr(limited.NominalConcurrencyShares))
		sections.AddText("Limit Response Type", string(limited.LimitResponse.Type))
		sections.AddText("Lendable Percent", formatInt32Ptr(limited.LendablePercent))
	}
	if exempt := config.Spec.Exempt; exempt != nil {
		sections.AddText("Exempt Lendable Percent", formatInt32Ptr(exempt.LendablePercent))
	}

	return component.NewSummary("Configuration", sections...), nil
}
