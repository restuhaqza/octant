/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// GatewayClassListHandler is a printFunc that prints gateway classes.
func GatewayClassListHandler(ctx context.Context, list *gatewayv1.GatewayClassList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("gateway class list is nil")
	}

	cols := component.NewTableCols("Name", "Controller", "Accepted", "Age")
	ot := NewObjectTable("Gateway Classes", "We couldn't find any gateway classes!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())

	for _, gatewayClass := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&gatewayClass, gatewayClass.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Controller"] = component.NewText(string(gatewayClass.Spec.ControllerName))
		row["Accepted"] = component.NewText(statusForCondition(gatewayClass.Status.Conditions, string(gatewayv1.GatewayClassConditionStatusAccepted)))
		row["Age"] = component.NewTimestamp(gatewayClass.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &gatewayClass, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// GatewayClassHandler is a printFunc that prints a single gateway class.
func GatewayClassHandler(ctx context.Context, gatewayClass *gatewayv1.GatewayClass, options Options) (component.Component, error) {
	obj := NewObject(gatewayClass)

	gch, err := newGatewayClassHandler(gatewayClass, obj)
	if err != nil {
		return nil, err
	}

	if err := gch.Config(options); err != nil {
		return nil, errors.Wrap(err, "print gateway class configuration")
	}

	return obj.ToComponent(ctx, options)
}

// GatewayClassConfiguration generates a gateway class configuration summary.
type GatewayClassConfiguration struct {
	gatewayClass *gatewayv1.GatewayClass
}

// NewGatewayClassConfiguration creates an instance of GatewayClassConfiguration.
func NewGatewayClassConfiguration(gatewayClass *gatewayv1.GatewayClass) *GatewayClassConfiguration {
	return &GatewayClassConfiguration{gatewayClass: gatewayClass}
}

// Create creates the configuration summary for a gateway class.
func (gcc *GatewayClassConfiguration) Create(options Options) (*component.Summary, error) {
	if gcc.gatewayClass == nil {
		return nil, errors.New("gateway class is nil")
	}

	gatewayClass := gcc.gatewayClass
	sections := component.SummarySections{}
	sections.AddText("Controller", string(gatewayClass.Spec.ControllerName))
	sections.AddText("Accepted", statusForCondition(gatewayClass.Status.Conditions, string(gatewayv1.GatewayClassConditionStatusAccepted)))

	if description := gatewayClass.Spec.Description; description != nil {
		sections.AddText("Description", *description)
	}

	if params := gatewayClass.Spec.ParametersRef; params != nil {
		sections.AddText("Parameters", fmt.Sprintf("%s/%s", params.Kind, params.Name))
	}

	return component.NewSummary("Configuration", sections...), nil
}

type gatewayClassHandler struct {
	configFunc   func(*gatewayv1.GatewayClass, Options) (*component.Summary, error)
	gatewayClass *gatewayv1.GatewayClass
	object       *Object
}

func newGatewayClassHandler(gatewayClass *gatewayv1.GatewayClass, object *Object) (*gatewayClassHandler, error) {
	if gatewayClass == nil {
		return nil, errors.New("can't print a nil gateway class")
	}
	if object == nil {
		return nil, errors.New("can't print gateway class using a nil object")
	}

	return &gatewayClassHandler{
		configFunc:   defaultGatewayClassConfig,
		gatewayClass: gatewayClass,
		object:       object,
	}, nil
}

func (gch *gatewayClassHandler) Config(options Options) error {
	out, err := gch.configFunc(gch.gatewayClass, options)
	if err != nil {
		return err
	}
	gch.object.RegisterConfig(out)
	return nil
}

func defaultGatewayClassConfig(gatewayClass *gatewayv1.GatewayClass, options Options) (*component.Summary, error) {
	return NewGatewayClassConfiguration(gatewayClass).Create(options)
}

// statusForCondition returns the status of the condition with the supplied type.
// It returns "Unknown" when the condition is absent.
func statusForCondition(conditions []metav1.Condition, conditionType string) string {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return string(conditions[i].Status)
		}
	}
	return "Unknown"
}
