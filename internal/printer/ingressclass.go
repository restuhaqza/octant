/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	networkingv1 "k8s.io/api/networking/v1"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// IngressClassListHandler is a printFunc that prints ingress classes.
func IngressClassListHandler(ctx context.Context, list *networkingv1.IngressClassList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("ingress class list is nil")
	}

	cols := component.NewTableCols("Name", "Controller", "Age")
	ot := NewObjectTable("Ingress Classes", "We couldn't find any ingress classes!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())
	for _, ingressClass := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&ingressClass, ingressClass.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Controller"] = component.NewText(ingressClass.Spec.Controller)
		row["Age"] = component.NewTimestamp(ingressClass.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &ingressClass, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// IngressClassHandler is a printFunc that prints a single ingress class.
func IngressClassHandler(ctx context.Context, ingressClass *networkingv1.IngressClass, options Options) (component.Component, error) {
	o := NewObject(ingressClass)
	o.EnableEvents()

	config, err := ingressClassConfig(ingressClass)
	if err != nil {
		return nil, err
	}
	o.RegisterConfig(config)

	return o.ToComponent(ctx, options)
}

func ingressClassConfig(ingressClass *networkingv1.IngressClass) (*component.Summary, error) {
	if ingressClass == nil {
		return nil, errors.New("ingress class is nil")
	}

	sections := component.SummarySections{}
	sections.AddText("Controller", ingressClass.Spec.Controller)
	if params := ingressClass.Spec.Parameters; params != nil {
		sections.AddText("Parameters Kind", params.Kind)
		sections.AddText("Parameters Name", params.Name)
		if params.APIGroup != nil {
			sections.AddText("Parameters API Group", *params.APIGroup)
		}
		if params.Scope != nil {
			sections.AddText("Parameters Scope", *params.Scope)
		}
		if params.Namespace != nil {
			sections.AddText("Parameters Namespace", *params.Namespace)
		}
	}

	return component.NewSummary("Configuration", sections...), nil
}
