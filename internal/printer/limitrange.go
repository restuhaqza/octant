/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// LimitRangeListHandler is a printFunc that prints limit ranges.
func LimitRangeListHandler(ctx context.Context, list *corev1.LimitRangeList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("limit range list is nil")
	}

	cols := component.NewTableCols("Name", "Limits", "Age")
	ot := NewObjectTable("Limit Ranges", "We couldn't find any limit ranges!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())
	for _, limitRange := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&limitRange, limitRange.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Limits"] = component.NewText(fmt.Sprintf("%d", len(limitRange.Spec.Limits)))
		row["Age"] = component.NewTimestamp(limitRange.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &limitRange, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// LimitRangeHandler is a printFunc that prints a single limit range.
func LimitRangeHandler(ctx context.Context, limitRange *corev1.LimitRange, options Options) (component.Component, error) {
	if limitRange == nil {
		return nil, errors.New("limit range is nil")
	}

	o := NewObject(limitRange)
	o.EnableEvents()

	o.RegisterItems(ItemDescriptor{
		Width: component.WidthFull,
		Func: func() (component.Component, error) {
			list := &corev1.LimitRangeList{Items: []corev1.LimitRange{*limitRange}}
			return printNamespaceResourceLimits(list)
		},
	})

	return o.ToComponent(ctx, options)
}
