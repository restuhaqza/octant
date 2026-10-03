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

	"github.com/vmware-tanzu/octant/internal/gvk"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// GatewayListHandler is a printFunc that prints gateways.
func GatewayListHandler(ctx context.Context, list *gatewayv1.GatewayList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("gateway list is nil")
	}

	cols := component.NewTableCols("Name", "Class", "Listeners", "Addresses", "Age")
	ot := NewObjectTable("Gateways", "We couldn't find any gateways!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())

	for _, gateway := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&gateway, gateway.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Class"] = component.NewText(string(gateway.Spec.GatewayClassName))
		row["Listeners"] = component.NewText(fmt.Sprintf("%d", len(gateway.Spec.Listeners)))
		row["Addresses"] = component.NewText(formatGatewayAddresses(gateway.Status.Addresses))
		row["Age"] = component.NewTimestamp(gateway.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &gateway, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// GatewayHandler is a printFunc that prints a single gateway.
func GatewayHandler(ctx context.Context, gateway *gatewayv1.Gateway, options Options) (component.Component, error) {
	obj := NewObject(gateway)

	gh, err := newGatewayHandler(gateway, obj)
	if err != nil {
		return nil, err
	}

	if err := gh.Config(options); err != nil {
		return nil, errors.Wrap(err, "print gateway configuration")
	}

	if err := gh.Listeners(options); err != nil {
		return nil, errors.Wrap(err, "print gateway listeners")
	}

	return obj.ToComponent(ctx, options)
}

// GatewayConfiguration generates a gateway configuration summary.
type GatewayConfiguration struct {
	gateway *gatewayv1.Gateway
}

// NewGatewayConfiguration creates an instance of GatewayConfiguration.
func NewGatewayConfiguration(gateway *gatewayv1.Gateway) *GatewayConfiguration {
	return &GatewayConfiguration{gateway: gateway}
}

// Create creates the configuration summary for a gateway.
func (gc *GatewayConfiguration) Create(options Options) (*component.Summary, error) {
	if gc.gateway == nil {
		return nil, errors.New("gateway is nil")
	}

	gateway := gc.gateway
	sections := component.SummarySections{}

	className := string(gateway.Spec.GatewayClassName)
	classLink, err := options.Link.ForGVK(
		"",
		gvk.GatewayClass.GroupVersion().String(),
		gvk.GatewayClass.Kind,
		className,
		className,
	)
	if err != nil {
		return nil, err
	}
	sections.Add("Gateway Class", classLink)

	sections.AddText("Listeners", fmt.Sprintf("%d", len(gateway.Spec.Listeners)))
	sections.AddText("Addresses", formatGatewayAddresses(gateway.Status.Addresses))
	sections.AddText("Accepted", statusForCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionAccepted)))
	sections.AddText("Programmed", statusForCondition(gateway.Status.Conditions, string(gatewayv1.GatewayConditionProgrammed)))

	return component.NewSummary("Configuration", sections...), nil
}

func createGatewayListenersView(gateway *gatewayv1.Gateway) (*component.Table, error) {
	if gateway == nil {
		return nil, errors.New("gateway is nil")
	}

	cols := component.NewTableCols("Name", "Hostname", "Port", "Protocol")
	table := component.NewTable("Listeners", "There are no listeners defined!", cols)

	for _, listener := range gateway.Spec.Listeners {
		hostname := "*"
		if listener.Hostname != nil {
			hostname = string(*listener.Hostname)
		}

		table.Add(component.TableRow{
			"Name":     component.NewText(string(listener.Name)),
			"Hostname": component.NewText(hostname),
			"Port":     component.NewText(fmt.Sprintf("%d", listener.Port)),
			"Protocol": component.NewText(string(listener.Protocol)),
		})
	}

	table.Sort("Name")

	return table, nil
}

func formatGatewayAddresses(addresses []gatewayv1.GatewayStatusAddress) string {
	var list []string
	for _, address := range addresses {
		list = append(list, address.Value)
	}
	if len(list) == 0 {
		return "<none>"
	}
	return strings.Join(list, ", ")
}

type gatewayHandler struct {
	gateway      *gatewayv1.Gateway
	configFunc   func(*gatewayv1.Gateway, Options) (*component.Summary, error)
	listenerFunc func(*gatewayv1.Gateway, Options) (component.Component, error)
	object       *Object
}

func newGatewayHandler(gateway *gatewayv1.Gateway, object *Object) (*gatewayHandler, error) {
	if gateway == nil {
		return nil, errors.New("can't print a nil gateway")
	}
	if object == nil {
		return nil, errors.New("can't print gateway using a nil object")
	}

	return &gatewayHandler{
		gateway:      gateway,
		configFunc:   defaultGatewayConfig,
		listenerFunc: defaultGatewayListeners,
		object:       object,
	}, nil
}

func (gh *gatewayHandler) Config(options Options) error {
	out, err := gh.configFunc(gh.gateway, options)
	if err != nil {
		return err
	}
	gh.object.RegisterConfig(out)
	return nil
}

func (gh *gatewayHandler) Listeners(options Options) error {
	if gh.gateway == nil {
		return errors.New("can't print listeners for nil gateway")
	}

	gh.object.RegisterItems(ItemDescriptor{
		Width: component.WidthFull,
		Func: func() (component.Component, error) {
			return gh.listenerFunc(gh.gateway, options)
		},
	})

	return nil
}

func defaultGatewayConfig(gateway *gatewayv1.Gateway, options Options) (*component.Summary, error) {
	return NewGatewayConfiguration(gateway).Create(options)
}

func defaultGatewayListeners(gateway *gatewayv1.Gateway, _ Options) (component.Component, error) {
	return createGatewayListenersView(gateway)
}
