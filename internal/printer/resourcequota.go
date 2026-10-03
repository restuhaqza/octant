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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/sets"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// ResourceQuotaListHandler is a printFunc that prints resource quotas.
func ResourceQuotaListHandler(ctx context.Context, list *corev1.ResourceQuotaList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("resource quota list is nil")
	}

	cols := component.NewTableCols("Name", "Scopes", "Age")
	ot := NewObjectTable("Resource Quotas", "We couldn't find any resource quotas!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())
	for _, quota := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&quota, quota.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Scopes"] = component.NewText(resourceQuotaScopes(&quota))
		row["Age"] = component.NewTimestamp(quota.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &quota, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// ResourceQuotaHandler is a printFunc that prints a single resource quota.
func ResourceQuotaHandler(ctx context.Context, quota *corev1.ResourceQuota, options Options) (component.Component, error) {
	o := NewObject(quota)
	o.EnableEvents()

	config, err := resourceQuotaConfig(quota)
	if err != nil {
		return nil, err
	}
	o.RegisterConfig(config)

	o.RegisterItems(ItemDescriptor{
		Width: component.WidthFull,
		Func: func() (component.Component, error) {
			return resourceQuotaResources(quota)
		},
	})

	return o.ToComponent(ctx, options)
}

func resourceQuotaConfig(quota *corev1.ResourceQuota) (*component.Summary, error) {
	if quota == nil {
		return nil, errors.New("resource quota is nil")
	}

	sections := component.SummarySections{}
	sections.AddText("Scopes", resourceQuotaScopes(quota))

	return component.NewSummary("Configuration", sections...), nil
}

func resourceQuotaScopes(quota *corev1.ResourceQuota) string {
	if quota == nil || len(quota.Spec.Scopes) == 0 {
		return "-"
	}

	scopes := make([]string, len(quota.Spec.Scopes))
	for i := range quota.Spec.Scopes {
		scopes[i] = string(quota.Spec.Scopes[i])
	}

	return strings.Join(scopes, ", ")
}

func resourceQuotaResources(quota *corev1.ResourceQuota) (*component.Table, error) {
	if quota == nil {
		return nil, errors.New("resource quota is nil")
	}

	cols := component.NewTableCols("Resource", "Used", "Hard")
	table := component.NewTable("Resources", "There are no resources!", cols)

	names := sets.NewString()
	for name := range quota.Status.Hard {
		names.Insert(name.String())
	}
	for name := range quota.Status.Used {
		names.Insert(name.String())
	}

	for _, name := range names.List() {
		resourceName := corev1.ResourceName(name)
		table.Add(component.TableRow{
			"Resource": component.NewText(name),
			"Used":     component.NewText(formatQuantityOrDash(quota.Status.Used[resourceName])),
			"Hard":     component.NewText(formatQuantityOrDash(quota.Status.Hard[resourceName])),
		})
	}

	table.Sort("Resource")

	return table, nil
}
