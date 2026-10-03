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

// FlowSchemaListHandler is a printFunc that prints flow schemas.
func FlowSchemaListHandler(ctx context.Context, list *flowcontrolv1.FlowSchemaList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("flow schema list is nil")
	}

	cols := component.NewTableCols("Name", "Priority Level", "Matching Precedence", "Age")
	ot := NewObjectTable("Flow Schemas", "We couldn't find any flow schemas!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())
	for _, flowSchema := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&flowSchema, flowSchema.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Priority Level"] = component.NewText(flowSchema.Spec.PriorityLevelConfiguration.Name)
		row["Matching Precedence"] = component.NewText(fmt.Sprintf("%d", flowSchema.Spec.MatchingPrecedence))
		row["Age"] = component.NewTimestamp(flowSchema.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &flowSchema, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// FlowSchemaHandler is a printFunc that prints a single flow schema.
func FlowSchemaHandler(ctx context.Context, flowSchema *flowcontrolv1.FlowSchema, options Options) (component.Component, error) {
	o := NewObject(flowSchema)
	o.EnableEvents()

	config, err := flowSchemaConfig(flowSchema)
	if err != nil {
		return nil, err
	}
	o.RegisterConfig(config)

	return o.ToComponent(ctx, options)
}

func flowSchemaConfig(flowSchema *flowcontrolv1.FlowSchema) (*component.Summary, error) {
	if flowSchema == nil {
		return nil, errors.New("flow schema is nil")
	}

	sections := component.SummarySections{}
	sections.AddText("Priority Level Configuration", flowSchema.Spec.PriorityLevelConfiguration.Name)
	sections.AddText("Matching Precedence", fmt.Sprintf("%d", flowSchema.Spec.MatchingPrecedence))
	if method := flowSchema.Spec.DistinguisherMethod; method != nil {
		sections.AddText("Distinguisher Method", string(method.Type))
	}
	sections.AddText("Rules", fmt.Sprintf("%d", len(flowSchema.Spec.Rules)))

	return component.NewSummary("Configuration", sections...), nil
}
