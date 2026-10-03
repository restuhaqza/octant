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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// HTTPRouteListHandler is a printFunc that prints http routes.
func HTTPRouteListHandler(ctx context.Context, list *gatewayv1.HTTPRouteList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("http route list is nil")
	}

	cols := component.NewTableCols("Name", "Hostnames", "Parents", "Backends", "Age")
	ot := NewObjectTable("HTTP Routes", "We couldn't find any http routes!", cols, options.DashConfig.ObjectStore())
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
		row["Backends"] = component.NewText(formatHTTPBackendRefs(route.Namespace, route.Spec.Rules))
		row["Age"] = component.NewTimestamp(route.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &route, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// HTTPRouteHandler is a printFunc that prints a single http route.
func HTTPRouteHandler(ctx context.Context, route *gatewayv1.HTTPRoute, options Options) (component.Component, error) {
	obj := NewObject(route)

	h, err := newHTTPRouteHandler(route, obj)
	if err != nil {
		return nil, err
	}

	if err := h.Config(options); err != nil {
		return nil, errors.Wrap(err, "print http route configuration")
	}

	if err := h.Rules(options); err != nil {
		return nil, errors.Wrap(err, "print http route rules")
	}

	return obj.ToComponent(ctx, options)
}

// HTTPRouteConfiguration generates an http route configuration summary.
type HTTPRouteConfiguration struct {
	route *gatewayv1.HTTPRoute
}

// NewHTTPRouteConfiguration creates an instance of HTTPRouteConfiguration.
func NewHTTPRouteConfiguration(route *gatewayv1.HTTPRoute) *HTTPRouteConfiguration {
	return &HTTPRouteConfiguration{route: route}
}

// Create creates the configuration summary for an http route.
func (c *HTTPRouteConfiguration) Create(options Options) (*component.Summary, error) {
	if c.route == nil {
		return nil, errors.New("http route is nil")
	}

	route := c.route
	sections := component.SummarySections{}
	sections.AddText("Hostnames", formatHostnames(route.Spec.Hostnames))
	sections.AddText("Parents", formatParentRefs(route.Namespace, route.Spec.ParentRefs))
	sections.AddText("Accepted", routeConditionsStatus(route.Status.RouteStatus, string(gatewayv1.RouteConditionAccepted)))
	sections.AddText("Resolved Refs", routeConditionsStatus(route.Status.RouteStatus, string(gatewayv1.RouteConditionResolvedRefs)))

	return component.NewSummary("Configuration", sections...), nil
}

func createHTTPRouteRulesView(route *gatewayv1.HTTPRoute) (*component.Table, error) {
	if route == nil {
		return nil, errors.New("http route is nil")
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
			"Matches":  component.NewText(formatHTTPRouteMatches(rule.Matches)),
			"Backends": component.NewText(formatHTTPRuleBackends(route.Namespace, rule.BackendRefs)),
		})
	}

	return table, nil
}

func formatHTTPRouteMatches(matches []gatewayv1.HTTPRouteMatch) string {
	if len(matches) == 0 {
		return "*"
	}

	var list []string
	for _, match := range matches {
		if match.Path != nil && match.Path.Value != nil {
			list = append(list, *match.Path.Value)
		}
	}
	if len(list) == 0 {
		return "*"
	}
	return strings.Join(list, ", ")
}

func formatHTTPBackendRefs(namespace string, rules []gatewayv1.HTTPRouteRule) string {
	var list []string
	for _, rule := range rules {
		if backends := formatHTTPRuleBackends(namespace, rule.BackendRefs); backends != "<none>" {
			list = append(list, backends)
		}
	}
	if len(list) == 0 {
		return "<none>"
	}
	return strings.Join(list, ", ")
}

func formatHTTPRuleBackends(namespace string, refs []gatewayv1.HTTPBackendRef) string {
	var list []string
	for _, ref := range refs {
		list = append(list, formatBackendObjectReference(namespace, ref.Group, ref.Kind, ref.Namespace, string(ref.Name)))
	}
	if len(list) == 0 {
		return "<none>"
	}
	return strings.Join(list, ", ")
}

func formatBackendObjectReference(defaultNamespace string, group *gatewayv1.Group, kind *gatewayv1.Kind, namespace *gatewayv1.Namespace, name string) string {
	objectKind := "Service"
	if kind != nil {
		objectKind = string(*kind)
	}
	if group != nil && *group != "" {
		objectKind = fmt.Sprintf("%s.%s", objectKind, *group)
	}

	ns := defaultNamespace
	if namespace != nil && *namespace != "" {
		ns = string(*namespace)
	}

	if ns == "" {
		return fmt.Sprintf("%s/%s", objectKind, name)
	}
	return fmt.Sprintf("%s %s/%s", objectKind, ns, name)
}

func formatHostnames(hostnames []gatewayv1.Hostname) string {
	if len(hostnames) == 0 {
		return "*"
	}

	list := make([]string, 0, len(hostnames))
	for _, hostname := range hostnames {
		list = append(list, string(hostname))
	}
	return strings.Join(list, ", ")
}

func formatParentRefs(defaultNamespace string, refs []gatewayv1.ParentReference) string {
	if len(refs) == 0 {
		return "<none>"
	}

	list := make([]string, 0, len(refs))
	for _, ref := range refs {
		kind := "Gateway"
		if ref.Kind != nil {
			kind = string(*ref.Kind)
		}

		ns := defaultNamespace
		if ref.Namespace != nil && *ref.Namespace != "" {
			ns = string(*ref.Namespace)
		}

		if ns == "" {
			list = append(list, fmt.Sprintf("%s/%s", kind, ref.Name))
		} else {
			list = append(list, fmt.Sprintf("%s %s/%s", kind, ns, ref.Name))
		}
	}
	return strings.Join(list, ", ")
}

// routeConditionsStatus aggregates a route condition type across all parents.
func routeConditionsStatus(routeStatus gatewayv1.RouteStatus, conditionType string) string {
	if len(routeStatus.Parents) == 0 {
		return "Unknown"
	}

	var found bool
	result := "True"
	for _, parent := range routeStatus.Parents {
		for i := range parent.Conditions {
			condition := parent.Conditions[i]
			if condition.Type != conditionType {
				continue
			}
			found = true
			if condition.Status == metav1.ConditionFalse {
				return "False"
			}
			if condition.Status == metav1.ConditionUnknown {
				result = "Unknown"
			}
		}
	}

	if !found {
		return "Unknown"
	}
	return result
}

type httpRouteHandler struct {
	route      *gatewayv1.HTTPRoute
	configFunc func(*gatewayv1.HTTPRoute, Options) (*component.Summary, error)
	rulesFunc  func(*gatewayv1.HTTPRoute, Options) (component.Component, error)
	object     *Object
}

func newHTTPRouteHandler(route *gatewayv1.HTTPRoute, object *Object) (*httpRouteHandler, error) {
	if route == nil {
		return nil, errors.New("can't print a nil http route")
	}
	if object == nil {
		return nil, errors.New("can't print http route using a nil object")
	}

	return &httpRouteHandler{
		route:      route,
		configFunc: defaultHTTPRouteConfig,
		rulesFunc:  defaultHTTPRouteRules,
		object:     object,
	}, nil
}

func (h *httpRouteHandler) Config(options Options) error {
	out, err := h.configFunc(h.route, options)
	if err != nil {
		return err
	}
	h.object.RegisterConfig(out)
	return nil
}

func (h *httpRouteHandler) Rules(options Options) error {
	if h.route == nil {
		return errors.New("can't print rules for nil http route")
	}

	h.object.RegisterItems(ItemDescriptor{
		Width: component.WidthFull,
		Func: func() (component.Component, error) {
			return h.rulesFunc(h.route, options)
		},
	})

	return nil
}

func defaultHTTPRouteConfig(route *gatewayv1.HTTPRoute, options Options) (*component.Summary, error) {
	return NewHTTPRouteConfiguration(route).Create(options)
}

func defaultHTTPRouteRules(route *gatewayv1.HTTPRoute, _ Options) (component.Component, error) {
	return createHTTPRouteRulesView(route)
}
