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
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// GRPCRouteListHandler is a printFunc that prints grpc routes.
func GRPCRouteListHandler(ctx context.Context, list *gatewayv1.GRPCRouteList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("grpc route list is nil")
	}

	cols := component.NewTableCols("Name", "Hostnames", "Parents", "Backends", "Age")
	ot := NewObjectTable("GRPC Routes", "We couldn't find any grpc routes!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())

	for _, route := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&route, route.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Hostnames"] = component.NewText(formatHostnames(route.Spec.Hostnames))
		row["Parents"] = component.NewText(formatParentRefs(route.Namespace, route.Spec.ParentRefs))
		row["Backends"] = component.NewText(formatGRPCBackendRefs(route.Namespace, route.Spec.Rules))
		row["Age"] = component.NewTimestamp(route.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &route, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// GRPCRouteHandler is a printFunc that prints a single grpc route.
func GRPCRouteHandler(ctx context.Context, route *gatewayv1.GRPCRoute, options Options) (component.Component, error) {
	obj := NewObject(route)

	h, err := newGRPCRouteHandler(route, obj)
	if err != nil {
		return nil, err
	}

	if err := h.Config(options); err != nil {
		return nil, errors.Wrap(err, "print grpc route configuration")
	}

	if err := h.Rules(options); err != nil {
		return nil, errors.Wrap(err, "print grpc route rules")
	}

	return obj.ToComponent(ctx, options)
}

// GRPCRouteConfiguration generates a grpc route configuration summary.
type GRPCRouteConfiguration struct {
	route *gatewayv1.GRPCRoute
}

// NewGRPCRouteConfiguration creates an instance of GRPCRouteConfiguration.
func NewGRPCRouteConfiguration(route *gatewayv1.GRPCRoute) *GRPCRouteConfiguration {
	return &GRPCRouteConfiguration{route: route}
}

// Create creates the configuration summary for a grpc route.
func (c *GRPCRouteConfiguration) Create(options Options) (*component.Summary, error) {
	if c.route == nil {
		return nil, errors.New("grpc route is nil")
	}

	route := c.route
	sections := component.SummarySections{}
	sections.AddText("Hostnames", formatHostnames(route.Spec.Hostnames))
	sections.AddText("Parents", formatParentRefs(route.Namespace, route.Spec.ParentRefs))
	sections.AddText("Accepted", routeConditionsStatus(route.Status.RouteStatus, string(gatewayv1.RouteConditionAccepted)))
	sections.AddText("Resolved Refs", routeConditionsStatus(route.Status.RouteStatus, string(gatewayv1.RouteConditionResolvedRefs)))

	return component.NewSummary("Configuration", sections...), nil
}

func createGRPCRouteRulesView(route *gatewayv1.GRPCRoute) (*component.Table, error) {
	if route == nil {
		return nil, errors.New("grpc route is nil")
	}

	cols := component.NewTableCols("Name", "Matches", "Backends")
	table := component.NewTable("Rules", "There are no rules defined!", cols)

	for i, rule := range route.Spec.Rules {
		name := fmt.Sprintf("rule-%d", i)
		if rule.Name != nil {
			name = string(*rule.Name)
		}

		table.Add(component.TableRow{
			"Name":     component.NewText(name),
			"Matches":  component.NewText(formatGRPCRouteMatches(rule.Matches)),
			"Backends": component.NewText(formatGRPCRuleBackends(route.Namespace, rule.BackendRefs)),
		})
	}

	return table, nil
}

func formatGRPCRouteMatches(matches []gatewayv1.GRPCRouteMatch) string {
	if len(matches) == 0 {
		return "*"
	}

	var list []string
	for _, match := range matches {
		if match.Method == nil {
			continue
		}
		service := "*"
		if match.Method.Service != nil && *match.Method.Service != "" {
			service = *match.Method.Service
		}
		method := "*"
		if match.Method.Method != nil && *match.Method.Method != "" {
			method = *match.Method.Method
		}
		list = append(list, fmt.Sprintf("%s/%s", service, method))
	}
	if len(list) == 0 {
		return "*"
	}
	return strings.Join(list, ", ")
}

func formatGRPCBackendRefs(namespace string, rules []gatewayv1.GRPCRouteRule) string {
	var list []string
	for _, rule := range rules {
		if backends := formatGRPCRuleBackends(namespace, rule.BackendRefs); backends != "<none>" {
			list = append(list, backends)
		}
	}
	if len(list) == 0 {
		return "<none>"
	}
	return strings.Join(list, ", ")
}

func formatGRPCRuleBackends(namespace string, refs []gatewayv1.GRPCBackendRef) string {
	var list []string
	for _, ref := range refs {
		list = append(list, formatBackendObjectReference(namespace, ref.Group, ref.Kind, ref.Namespace, string(ref.Name)))
	}
	if len(list) == 0 {
		return "<none>"
	}
	return strings.Join(list, ", ")
}

type grpcRouteHandler struct {
	route      *gatewayv1.GRPCRoute
	configFunc func(*gatewayv1.GRPCRoute, Options) (*component.Summary, error)
	rulesFunc  func(*gatewayv1.GRPCRoute, Options) (component.Component, error)
	object     *Object
}

func newGRPCRouteHandler(route *gatewayv1.GRPCRoute, object *Object) (*grpcRouteHandler, error) {
	if route == nil {
		return nil, errors.New("can't print a nil grpc route")
	}
	if object == nil {
		return nil, errors.New("can't print grpc route using a nil object")
	}

	return &grpcRouteHandler{
		route:      route,
		configFunc: defaultGRPCRouteConfig,
		rulesFunc:  defaultGRPCRouteRules,
		object:     object,
	}, nil
}

func (h *grpcRouteHandler) Config(options Options) error {
	out, err := h.configFunc(h.route, options)
	if err != nil {
		return err
	}
	h.object.RegisterConfig(out)
	return nil
}

func (h *grpcRouteHandler) Rules(options Options) error {
	if h.route == nil {
		return errors.New("can't print rules for nil grpc route")
	}

	h.object.RegisterItems(ItemDescriptor{
		Width: component.WidthFull,
		Func: func() (component.Component, error) {
			return h.rulesFunc(h.route, options)
		},
	})

	return nil
}

func defaultGRPCRouteConfig(route *gatewayv1.GRPCRoute, options Options) (*component.Summary, error) {
	return NewGRPCRouteConfiguration(route).Create(options)
}

func defaultGRPCRouteRules(route *gatewayv1.GRPCRoute, _ Options) (component.Component, error) {
	return createGRPCRouteRulesView(route)
}
