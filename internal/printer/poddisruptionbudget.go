/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	policyv1 "k8s.io/api/policy/v1"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// PodDisruptionBudgetListHandler is a printFunc that prints pod disruption budgets.
func PodDisruptionBudgetListHandler(ctx context.Context, list *policyv1.PodDisruptionBudgetList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("pod disruption budget list is nil")
	}

	cols := component.NewTableCols("Name", "Min Available", "Max Unavailable", "Allowed Disruptions", "Age")
	ot := NewObjectTable("Pod Disruption Budgets", "We couldn't find any pod disruption budgets!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())
	for _, pdb := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&pdb, pdb.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Min Available"] = component.NewText(formatIntOrString(pdb.Spec.MinAvailable))
		row["Max Unavailable"] = component.NewText(formatIntOrString(pdb.Spec.MaxUnavailable))
		row["Allowed Disruptions"] = component.NewText(fmt.Sprintf("%d", pdb.Status.DisruptionsAllowed))
		row["Age"] = component.NewTimestamp(pdb.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &pdb, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// PodDisruptionBudgetHandler is a printFunc that prints a single pod disruption budget.
func PodDisruptionBudgetHandler(ctx context.Context, pdb *policyv1.PodDisruptionBudget, options Options) (component.Component, error) {
	o := NewObject(pdb)
	o.EnableEvents()

	config, err := podDisruptionBudgetConfig(pdb)
	if err != nil {
		return nil, err
	}
	o.RegisterConfig(config)

	status := podDisruptionBudgetStatus(pdb)
	o.RegisterSummary(status)

	return o.ToComponent(ctx, options)
}

func podDisruptionBudgetConfig(pdb *policyv1.PodDisruptionBudget) (*component.Summary, error) {
	if pdb == nil {
		return nil, errors.New("pod disruption budget is nil")
	}

	sections := component.SummarySections{}
	sections.AddText("Min Available", formatIntOrString(pdb.Spec.MinAvailable))
	sections.AddText("Max Unavailable", formatIntOrString(pdb.Spec.MaxUnavailable))
	sections.Add("Selector", printSelector(pdb.Spec.Selector))
	if pdb.Spec.UnhealthyPodEvictionPolicy != nil {
		sections.AddText("Unhealthy Pod Eviction Policy", string(*pdb.Spec.UnhealthyPodEvictionPolicy))
	}

	return component.NewSummary("Configuration", sections...), nil
}

func podDisruptionBudgetStatus(pdb *policyv1.PodDisruptionBudget) *component.Summary {
	if pdb == nil {
		return nil
	}

	sections := component.SummarySections{}
	sections.AddText("Current Healthy", fmt.Sprintf("%d", pdb.Status.CurrentHealthy))
	sections.AddText("Desired Healthy", fmt.Sprintf("%d", pdb.Status.DesiredHealthy))
	sections.AddText("Expected Pods", fmt.Sprintf("%d", pdb.Status.ExpectedPods))
	sections.AddText("Disruptions Allowed", fmt.Sprintf("%d", pdb.Status.DisruptionsAllowed))

	return component.NewSummary("Status", sections...)
}
