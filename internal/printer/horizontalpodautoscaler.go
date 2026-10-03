/*
Copyright (c) 2020 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/vmware-tanzu/octant/internal/util/json"

	"github.com/pkg/errors"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/octant/pkg/store"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// HorizontalPodAutoscalerListHandler is a printFunc that lists horizontal pod autoscalers
func HorizontalPodAutoscalerListHandler(ctx context.Context, list *autoscalingv1.HorizontalPodAutoscalerList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("horizontalpod handler list is nil")
	}

	cols := component.NewTableCols("Name", "Labels", "Targets", "Minimum Pods", "Maximum Pods", "Replicas", "Age")
	ot := NewObjectTable("Horizontal Pod Autoscalers",
		"We couldn't find any horizontal pod autoscalers", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())
	for _, horizontalPodAutoscaler := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&horizontalPodAutoscaler, horizontalPodAutoscaler.Name)
		if err != nil {
			return nil, err
		}

		horizontalPodAutoscalerMetrics, horizontalPodAutoscalerCurrentMetrics, err := parseAnnotations(horizontalPodAutoscaler)
		if err != nil {
			return nil, errors.Wrap(err, "can't parse annotations")
		}

		aggregatedMetricTargets, err := getCombinedMetrics(horizontalPodAutoscaler, horizontalPodAutoscalerMetrics, horizontalPodAutoscalerCurrentMetrics)
		if err != nil {
			return nil, errors.Wrap(err, "can't combine metrics")
		}

		row["Name"] = nameLink
		row["Labels"] = component.NewLabels(horizontalPodAutoscaler.Labels)
		row["Targets"] = component.NewText(aggregatedMetricTargets)
		row["Minimum Pods"] = component.NewText(fmt.Sprintf("%d", *horizontalPodAutoscaler.Spec.MinReplicas))
		row["Maximum Pods"] = component.NewText(fmt.Sprintf("%d", horizontalPodAutoscaler.Spec.MaxReplicas))
		row["Replicas"] = component.NewText(fmt.Sprintf("%d", horizontalPodAutoscaler.Status.CurrentReplicas))
		row["Age"] = component.NewTimestamp(horizontalPodAutoscaler.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &horizontalPodAutoscaler, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// HorizontalPodAutoscalerHandler is a printFunc that prints a HorizontalPodAutoscaler
func HorizontalPodAutoscalerHandler(ctx context.Context, horizontalPodAutoscaler *autoscalingv1.HorizontalPodAutoscaler, options Options) (component.Component, error) {
	o := NewObject(horizontalPodAutoscaler)
	o.EnableEvents()
	o.DisableConditions()

	hh, err := newHorizontalPodAutoscalerHandler(horizontalPodAutoscaler, o)
	if err != nil {
		return nil, err
	}

	if err := hh.Config(ctx, options); err != nil {
		return nil, errors.Wrap(err, "print horizontalpodautoscaler configuration")
	}

	if err := hh.Status(); err != nil {
		return nil, errors.Wrap(err, "print horizontalpodautoscaler status")
	}

	if err := hh.Metrics(ctx, options); err != nil {
		return nil, errors.Wrap(err, "print horizontalpodautoscaler metrics")
	}

	if err := hh.Conditions(); err != nil {
		return nil, errors.Wrap(err, "print horizontalpodautoscaler conditions")
	}

	return o.ToComponent(ctx, options)
}

func createHorizontalPodAutoscalerSummaryStatus(horizontalPodAutoscaler *autoscalingv1.HorizontalPodAutoscaler) (*component.Summary, error) {
	if horizontalPodAutoscaler == nil {
		return nil, errors.New("unable to generate status for a nil horizontalpodautoscaler")
	}

	horizontalPodAutoscalerMetrics, horizontalPodAutoscalerCurrentMetrics, err := parseAnnotations(*horizontalPodAutoscaler)
	if err != nil {
		return nil, errors.Wrap(err, "can't parse annotations")
	}

	aggregatedMetricTargets, err := getCombinedMetrics(*horizontalPodAutoscaler, horizontalPodAutoscalerMetrics, horizontalPodAutoscalerCurrentMetrics)
	if err != nil {
		return nil, errors.Wrap(err, "can't combine metrics")
	}

	status := horizontalPodAutoscaler.Status

	summary := component.NewSummary("Status")

	sections := component.SummarySections{}

	sections.AddText("Targets", aggregatedMetricTargets)

	if status.ObservedGeneration != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Observed Generation",
			Content: component.NewText(fmt.Sprintf("%d", *status.ObservedGeneration)),
		})
	}

	if status.LastScaleTime != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Last Scale Time",
			Content: component.NewTimestamp(status.LastScaleTime.Time),
		})
	}

	sections.AddText("Current Replicas", fmt.Sprintf("%d", status.CurrentReplicas))
	sections.AddText("Desired Replicas", fmt.Sprintf("%d", status.DesiredReplicas))

	if status.CurrentCPUUtilizationPercentage != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Current CPU Utilization Percentage",
			Content: component.NewText(fmt.Sprintf("%d", *status.CurrentCPUUtilizationPercentage)),
		})
	}

	summary.Add(sections...)

	return summary, nil
}

func createHorizontalPodAutoscalerMetricsStatusView(metricStatus *autoscalingv1.MetricStatus, options Options) (*component.Summary, error) {
	if metricStatus == nil {
		return nil, errors.New("unable to generate metrics from a nil metric status")
	}

	sections := component.SummarySections{}
	summary := component.NewSummary(fmt.Sprintf("Metric"))

	sections.AddText("Type", fmt.Sprintf("%s", metricStatus.Type))

	if metricStatus.Object != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Name",
			Content: component.NewText(metricStatus.Object.MetricName),
		})

		if metricStatus.Object.Target.Name != "" {
			sections = append(sections, component.SummarySection{
				Header:  "Described Object Name",
				Content: component.NewText(metricStatus.Object.Target.Name),
			})
			sections = append(sections, component.SummarySection{
				Header:  "Described Object API Version",
				Content: component.NewText(metricStatus.Object.Target.APIVersion),
			})
			sections = append(sections, component.SummarySection{
				Header:  "Described Object Kind",
				Content: component.NewText(metricStatus.Object.Target.Kind),
			})
		}
	}

	if metricStatus.Pods != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Name",
			Content: component.NewText(metricStatus.Pods.MetricName),
		})
		sections = append(sections, component.SummarySection{
			Header:  "Average Utilization",
			Content: component.NewText(fmt.Sprint(&metricStatus.Pods.CurrentAverageValue)),
		})
	}

	if metricStatus.Resource != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Name",
			Content: component.NewText(string(metricStatus.Resource.Name)),
		})
		if metricStatus.Resource.CurrentAverageUtilization != nil {
			sections = append(sections, component.SummarySection{
				Header:  "Average Utilization",
				Content: component.NewText(fmt.Sprint(*metricStatus.Resource.CurrentAverageUtilization)),
			})
		}
		sections = append(sections, component.SummarySection{
			Header:  "Average Value",
			Content: component.NewText(metricStatus.Resource.CurrentAverageValue.String()),
		})
	}

	summary.Add(sections...)

	return summary, nil
}

var hpaConditionColumns = [][]string{
	{"Type", "type"},
	{"Reason", "reason"},
	{"Status", "status"},
	{"Message", "message"},
	{"Last Transition", "lastTransitionTime"},
}

func createHorizontalPodAutoscalerConditionsView(horizontalPodAutoscaler *autoscalingv1.HorizontalPodAutoscaler) (*component.Table, error) {
	horizontalPodAutoscalerConditions := make([]interface{}, 0)

	if conditions, ok := horizontalPodAutoscaler.Annotations["autoscaling.alpha.kubernetes.io/conditions"]; ok {
		err := json.Unmarshal([]byte(conditions), &horizontalPodAutoscalerConditions)
		if err != nil {
			return nil, err
		}
	}

	object := map[string]interface{}{
		"status": map[string]interface{}{
			"conditions": horizontalPodAutoscalerConditions,
		},
	}

	conditions, err := parseConditions(unstructured.Unstructured{Object: object})
	if err != nil {
		return nil, err
	}
	table := createConditionsTable(conditions, conditionType, hpaConditionColumns)
	return table, nil
}

// HorizontalPodAutoscalerConfiguration generates a horizontalpodautoscaler configuration
type HorizontalPodAutoscalerConfiguration struct {
	horizontalPodAutoscaler *autoscalingv1.HorizontalPodAutoscaler
}

// NewHorizontalPodAutoscalerConfiguration creates an instance of HorizontalPodAutoscalerConfiguration
func NewHorizontalPodAutoscalerConfiguration(hpa *autoscalingv1.HorizontalPodAutoscaler) *HorizontalPodAutoscalerConfiguration {
	return &HorizontalPodAutoscalerConfiguration{
		horizontalPodAutoscaler: hpa,
	}
}

type horizontalPodAutoscalerObject interface {
	Config(ctx context.Context, options Options) error
	Status() error
	Metrics(ctx context.Context, options Options) error
	Conditions() error
}

type horizontalPodAutoscalerHandler struct {
	horizontalPodAutoScaler *autoscalingv1.HorizontalPodAutoscaler
	configFunc              func(context.Context, *autoscalingv1.HorizontalPodAutoscaler, Options) (*component.Summary, error)
	statusFunc              func(*autoscalingv1.HorizontalPodAutoscaler) (*component.Summary, error)
	metricsFunc             func(context.Context, *autoscalingv1.MetricStatus, Options) (*component.Summary, error)
	conditionsFunc          func(*autoscalingv1.HorizontalPodAutoscaler) (*component.Table, error)
	object                  *Object
}

// Create creates a horizontalpodautoscaler configuration summary
func (hc *HorizontalPodAutoscalerConfiguration) Create(ctx context.Context, options Options) (*component.Summary, error) {
	if hc.horizontalPodAutoscaler == nil {
		return nil, errors.New("horizontalpodautoscaler is nil")
	}

	hpa := hc.horizontalPodAutoscaler

	sections := component.SummarySections{}

	scaleTarget, err := forScaleTarget(ctx, hpa, &hpa.Spec.ScaleTargetRef, options)
	if err != nil {
		return nil, err
	}

	sections = append(sections, component.SummarySection{
		Header:  "Scale target",
		Content: scaleTarget,
	})

	minReplicas := fmt.Sprintf("%d", *hpa.Spec.MinReplicas)
	maxReplicas := fmt.Sprintf("%d", hpa.Spec.MaxReplicas)
	sections.AddText("Min Replicas", minReplicas)
	sections.AddText("Max Replicas", maxReplicas)

	b := autoscalingv2.HorizontalPodAutoscalerBehavior{}

	if behavior, ok := hpa.Annotations["autoscaling.alpha.kubernetes.io/behavior"]; ok {
		if err := json.Unmarshal([]byte(behavior), &b); err != nil {
			return nil, err
		}

		sections = append(sections, hpaBehaviorSections(&b)...)
	}

	summary := component.NewSummary("Configuration", sections...)
	return summary, nil
}

// hpaBehaviorSections renders the HPA scaling behavior summary sections. It is shared by the
// v1 annotation-derived behavior and the native autoscaling/v2 spec.behavior field.
func hpaBehaviorSections(behavior *autoscalingv2.HorizontalPodAutoscalerBehavior) []component.SummarySection {
	if behavior == nil {
		return nil
	}

	var sections []component.SummarySection

	if behavior.ScaleUp != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Scale Up",
			Content: hpaScalingRulesTable(behavior.ScaleUp, "There are no scale up policies!"),
		})
	}

	if behavior.ScaleDown != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Scale Down",
			Content: hpaScalingRulesTable(behavior.ScaleDown, "There are no scale down policies!"),
		})
	}

	return sections
}

func hpaScalingRulesTable(rules *autoscalingv2.HPAScalingRules, emptyMessage string) *component.Table {
	var policies []string
	for _, policy := range rules.Policies {
		policies = append(policies, fmt.Sprintf("%d %s / %d seconds", policy.Value, policy.Type, policy.PeriodSeconds))
	}

	stabilizationWindow := ""
	if rules.StabilizationWindowSeconds != nil {
		stabilizationWindow = fmt.Sprint(*rules.StabilizationWindowSeconds) + " seconds"
	}

	selectPolicy := ""
	if rules.SelectPolicy != nil {
		selectPolicy = fmt.Sprint(*rules.SelectPolicy)
	}

	cols := component.NewTableCols("Stabilization Window", "Select Policies", "Policies")
	return component.NewTableWithRows("", emptyMessage, cols, []component.TableRow{
		{
			"Stabilization Window": component.NewText(stabilizationWindow),
			"Select Policies":      component.NewText(selectPolicy),
			"Policies":             component.NewText(strings.Join(policies, ", ")),
		},
	})
}

var _ horizontalPodAutoscalerObject = (*horizontalPodAutoscalerHandler)(nil)

func newHorizontalPodAutoscalerHandler(horizontalPodAutoscaler *autoscalingv1.HorizontalPodAutoscaler, object *Object) (*horizontalPodAutoscalerHandler, error) {
	if horizontalPodAutoscaler == nil {
		return nil, errors.New("can't print a nil horizontalpodautoscaler")
	}

	if object == nil {
		return nil, errors.New("can't print horizontalpodautoscaler using a nil object printer")
	}

	hh := &horizontalPodAutoscalerHandler{
		horizontalPodAutoScaler: horizontalPodAutoscaler,
		configFunc:              defaultHorizontalPodAutoscalerConfig,
		statusFunc:              defaultHorizontalPodAutoscalerStatus,
		metricsFunc:             defaultHorizontalPodAutoscalerMetrics,
		conditionsFunc:          defaultHorizontalPodAutoscalerConditions,
		object:                  object,
	}

	return hh, nil
}

func (h *horizontalPodAutoscalerHandler) Config(ctx context.Context, options Options) error {
	out, err := h.configFunc(ctx, h.horizontalPodAutoScaler, options)
	if err != nil {
		return err
	}

	h.object.RegisterConfig(out)
	return nil
}

func defaultHorizontalPodAutoscalerConfig(ctx context.Context, horizontalPodAutoscaler *autoscalingv1.HorizontalPodAutoscaler, options Options) (*component.Summary, error) {
	return NewHorizontalPodAutoscalerConfiguration(horizontalPodAutoscaler).Create(ctx, options)
}

func (h *horizontalPodAutoscalerHandler) Status() error {
	out, err := h.statusFunc(h.horizontalPodAutoScaler)
	if err != nil {
		return err
	}

	h.object.RegisterSummary(out)
	return nil
}

func defaultHorizontalPodAutoscalerStatus(horizontalPodAutoscaler *autoscalingv1.HorizontalPodAutoscaler) (*component.Summary, error) {
	return createHorizontalPodAutoscalerSummaryStatus(horizontalPodAutoscaler)
}

func (h *horizontalPodAutoscalerHandler) metrics(ctx context.Context, currentMetrics []autoscalingv1.MetricStatus, options Options) error {
	if h == nil || h.horizontalPodAutoScaler == nil {
		return errors.New("can't display metrics for nil horizontalpodautoscaler")
	}

	for i := range currentMetrics {
		metric := currentMetrics[i]

		if metric.Type == "" {
			continue
		}

		h.object.RegisterItems(ItemDescriptor{
			Width: component.WidthFull,
			Func: func() (component.Component, error) {
				return h.metricsFunc(ctx, &metric, options)
			},
		})
	}

	return nil
}

func (h *horizontalPodAutoscalerHandler) Metrics(ctx context.Context, options Options) error {
	if h.horizontalPodAutoScaler == nil {
		return errors.New("can't display metrics for nil horizontalpodautoscaler")
	}

	_, metricStatus, err := parseAnnotations(*h.horizontalPodAutoScaler)
	if err != nil {
		return errors.New("can't parse annotations for metrics")
	}

	return h.metrics(ctx, metricStatus, options)
}

func defaultHorizontalPodAutoscalerMetrics(ctx context.Context, metricStatus *autoscalingv1.MetricStatus, options Options) (*component.Summary, error) {
	return createHorizontalPodAutoscalerMetricsStatusView(metricStatus, options)
}

func (h *horizontalPodAutoscalerHandler) Conditions() error {
	if h.horizontalPodAutoScaler == nil {
		return errors.New("can't display conditions for nil horizontalpodautoscaler")
	}

	h.object.RegisterItems(ItemDescriptor{
		Width: component.WidthFull,
		Func: func() (component.Component, error) {
			return h.conditionsFunc(h.horizontalPodAutoScaler)
		},
	})

	return nil
}

func defaultHorizontalPodAutoscalerConditions(horizontalPodAutoscaler *autoscalingv1.HorizontalPodAutoscaler) (*component.Table, error) {
	return createHorizontalPodAutoscalerConditionsView(horizontalPodAutoscaler)
}

// forScaleTarget returns a scale target for a cross version object reference
func forScaleTarget(ctx context.Context, object runtime.Object, scaleTarget *autoscalingv1.CrossVersionObjectReference, options Options) (*component.Link, error) {
	if scaleTarget == nil {
		return forScaleTargetRef(ctx, object, "", "", "", options)
	}

	return forScaleTargetRef(ctx, object, scaleTarget.APIVersion, scaleTarget.Kind, scaleTarget.Name, options)
}

// forScaleTargetRef returns a scale target link for a version-agnostic cross version object reference.
func forScaleTargetRef(ctx context.Context, object runtime.Object, apiVersion, kind, name string, options Options) (*component.Link, error) {
	if object == nil || (apiVersion == "" && kind == "" && name == "") {
		return component.NewLink("", "none", ""), nil
	}

	accessor := meta.NewAccessor()
	ns, err := accessor.Namespace(object)
	if err != nil {
		return component.NewLink("", "none", ""), nil
	}

	key := store.Key{
		Namespace:  ns,
		APIVersion: apiVersion,
		Kind:       kind,
		Name:       name,
	}

	objectStore := options.DashConfig.ObjectStore()
	u, err := objectStore.Get(ctx, key)
	if err != nil || u == nil {
		return component.NewLink("", "none", ""), nil
	}

	return options.Link.ForGVK(
		ns,
		apiVersion,
		kind,
		name,
		name,
	)
}

func getCombinedMetrics(horizontalPodAutoscaler autoscalingv1.HorizontalPodAutoscaler, metricSpec []autoscalingv1.MetricSpec, metricStatus []autoscalingv1.MetricStatus) (string, error) {
	var targets = make(map[string]string)
	var currents = make(map[string]string)

	for _, m := range metricSpec {
		switch m.Type {
		case autoscalingv1.ObjectMetricSourceType:
			if m.Object != nil {
				if m.Object.MetricName != "" && &m.Object.TargetValue != nil {
					targets[m.Object.MetricName] = m.Object.TargetValue.String()
				}
			}
		case autoscalingv1.PodsMetricSourceType:
			if m.Pods != nil {
				if m.Pods.MetricName != "" && &m.Pods.TargetAverageValue != nil {
					target := m.Pods.TargetAverageValue.String()
					targets[m.Pods.MetricName] = target
				}
			}

		case autoscalingv1.ResourceMetricSourceType:
			if m.Resource != nil {
				if m.Resource.Name != "" && m.Resource.TargetAverageValue != nil {
					targets[string(m.Resource.Name)] = m.Resource.TargetAverageValue.String()
				}
			}

		case autoscalingv1.ExternalMetricSourceType:
			if m.External != nil {
				if m.External.MetricName != "" && m.External.TargetAverageValue != nil {
					targets[string(m.External.MetricName)] = m.External.TargetAverageValue.String()
				}
			}
		}
	}

	for _, m := range metricStatus {
		current, err := getMetricStatusValue(&m)
		if err != nil {
			return "", err
		}

		switch m.Type {
		case autoscalingv1.ObjectMetricSourceType:
			if m.Object != nil {
				if m.Object.MetricName != "" {
					currents[m.Object.MetricName] = current
				}
			}
		case autoscalingv1.PodsMetricSourceType:
			if m.Pods != nil {
				if m.Pods.MetricName != "" {
					currents[m.Pods.MetricName] = current
				}
			}
		case autoscalingv1.ResourceMetricSourceType:
			if m.Resource != nil {
				if m.Resource.Name != "" {
					currents[string(m.Resource.Name)] = current
				}
			}
		case autoscalingv1.ExternalMetricSourceType:
			if m.External != nil {
				if m.External.MetricName != "" {
					currents[string(m.External.MetricName)] = current
				}
			}
		}
	}

	var result []string
	keys := make([]string, 0, len(targets))
	for k := range targets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if currents[k] == "" {
			currents[k] = "<unknown>"
		}
		result = append(result, currents[k]+"/"+targets[k])
	}

	if horizontalPodAutoscaler.Spec.TargetCPUUtilizationPercentage != nil && horizontalPodAutoscaler.Status.CurrentCPUUtilizationPercentage != nil {
		cpu := fmt.Sprintf("%d/%d", *horizontalPodAutoscaler.Status.CurrentCPUUtilizationPercentage, *horizontalPodAutoscaler.Spec.TargetCPUUtilizationPercentage) + "%"
		result = append(result, cpu)
	}

	return strings.Join(result, ", "), nil
}

func getMetricStatusValue(metricStatus *autoscalingv1.MetricStatus) (string, error) {
	var value string
	if metricStatus == nil {
		return "", errors.New("nil metric status")
	}

	switch metricStatus.Type {
	case autoscalingv1.ObjectMetricSourceType:
		if &metricStatus.Object.CurrentValue != nil {
			value = metricStatus.Object.CurrentValue.String()
		}
	case autoscalingv1.PodsMetricSourceType:
		if &metricStatus.Pods.CurrentAverageValue != nil {
			value = metricStatus.Pods.CurrentAverageValue.String()
		}
	case autoscalingv1.ResourceMetricSourceType:
		if &metricStatus.Resource.CurrentAverageValue != nil {
			value = metricStatus.Resource.CurrentAverageValue.String()
		}
	case autoscalingv1.ExternalMetricSourceType:
		if metricStatus.External.CurrentAverageValue != nil {
			value = metricStatus.External.CurrentAverageValue.String()
		}
	}

	return value, nil
}

func parseAnnotations(horizontalPodAutoscaler autoscalingv1.HorizontalPodAutoscaler) ([]autoscalingv1.MetricSpec, []autoscalingv1.MetricStatus, error) {
	horizontalPodAutoscalerMetrics := make([]autoscalingv1.MetricSpec, 0)
	horizontalPodAutoscalerCurrentMetrics := make([]autoscalingv1.MetricStatus, 0)

	if metrics, ok := horizontalPodAutoscaler.Annotations["autoscaling.alpha.kubernetes.io/metrics"]; ok {
		err := json.Unmarshal([]byte(metrics), &horizontalPodAutoscalerMetrics)
		if err != nil {
			return nil, nil, err
		}
	}
	if currentMetrics, ok := horizontalPodAutoscaler.Annotations["autoscaling.alpha.kubernetes.io/current-metrics"]; ok {
		err := json.Unmarshal([]byte(currentMetrics), &horizontalPodAutoscalerCurrentMetrics)
		if err != nil {
			return nil, nil, err
		}
	}

	return horizontalPodAutoscalerMetrics, horizontalPodAutoscalerCurrentMetrics, nil
}

// HorizontalPodAutoscalerV2ListHandler is a printFunc that lists horizontal pod autoscalers for autoscaling/v2.
func HorizontalPodAutoscalerV2ListHandler(ctx context.Context, list *autoscalingv2.HorizontalPodAutoscalerList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("horizontalpod handler list is nil")
	}

	cols := component.NewTableCols("Name", "Labels", "Targets", "Minimum Pods", "Maximum Pods", "Replicas", "Age")
	ot := NewObjectTable("Horizontal Pod Autoscalers",
		"We couldn't find any horizontal pod autoscalers", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())
	for i := range list.Items {
		horizontalPodAutoscaler := &list.Items[i]
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(horizontalPodAutoscaler, horizontalPodAutoscaler.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Labels"] = component.NewLabels(horizontalPodAutoscaler.Labels)
		row["Targets"] = component.NewText(getCombinedMetricsV2(horizontalPodAutoscaler))
		row["Minimum Pods"] = component.NewText(fmt.Sprintf("%d", minReplicasV2(horizontalPodAutoscaler)))
		row["Maximum Pods"] = component.NewText(fmt.Sprintf("%d", horizontalPodAutoscaler.Spec.MaxReplicas))
		row["Replicas"] = component.NewText(fmt.Sprintf("%d", horizontalPodAutoscaler.Status.CurrentReplicas))
		row["Age"] = component.NewTimestamp(horizontalPodAutoscaler.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, horizontalPodAutoscaler, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// HorizontalPodAutoscalerV2Handler is a printFunc that prints an autoscaling/v2 HorizontalPodAutoscaler.
func HorizontalPodAutoscalerV2Handler(ctx context.Context, horizontalPodAutoscaler *autoscalingv2.HorizontalPodAutoscaler, options Options) (component.Component, error) {
	o := NewObject(horizontalPodAutoscaler)
	o.EnableEvents()
	o.DisableConditions()

	hh, err := newHorizontalPodAutoscalerV2Handler(horizontalPodAutoscaler, o)
	if err != nil {
		return nil, err
	}

	if err := hh.Config(ctx, options); err != nil {
		return nil, errors.Wrap(err, "print horizontalpodautoscaler configuration")
	}

	if err := hh.Status(); err != nil {
		return nil, errors.Wrap(err, "print horizontalpodautoscaler status")
	}

	if err := hh.Metrics(ctx, options); err != nil {
		return nil, errors.Wrap(err, "print horizontalpodautoscaler metrics")
	}

	if err := hh.Conditions(); err != nil {
		return nil, errors.Wrap(err, "print horizontalpodautoscaler conditions")
	}

	return o.ToComponent(ctx, options)
}

// HorizontalPodAutoscalerV2Configuration generates an autoscaling/v2 horizontalpodautoscaler configuration.
type HorizontalPodAutoscalerV2Configuration struct {
	horizontalPodAutoscaler *autoscalingv2.HorizontalPodAutoscaler
}

// NewHorizontalPodAutoscalerV2Configuration creates an instance of HorizontalPodAutoscalerV2Configuration.
func NewHorizontalPodAutoscalerV2Configuration(hpa *autoscalingv2.HorizontalPodAutoscaler) *HorizontalPodAutoscalerV2Configuration {
	return &HorizontalPodAutoscalerV2Configuration{
		horizontalPodAutoscaler: hpa,
	}
}

// Create creates an autoscaling/v2 horizontalpodautoscaler configuration summary.
func (hc *HorizontalPodAutoscalerV2Configuration) Create(ctx context.Context, options Options) (*component.Summary, error) {
	if hc.horizontalPodAutoscaler == nil {
		return nil, errors.New("horizontalpodautoscaler is nil")
	}

	hpa := hc.horizontalPodAutoscaler

	sections := component.SummarySections{}

	scaleTarget, err := forScaleTargetRef(ctx, hpa, hpa.Spec.ScaleTargetRef.APIVersion, hpa.Spec.ScaleTargetRef.Kind, hpa.Spec.ScaleTargetRef.Name, options)
	if err != nil {
		return nil, err
	}

	sections = append(sections, component.SummarySection{
		Header:  "Scale target",
		Content: scaleTarget,
	})

	sections.AddText("Min Replicas", fmt.Sprintf("%d", minReplicasV2(hpa)))
	sections.AddText("Max Replicas", fmt.Sprintf("%d", hpa.Spec.MaxReplicas))

	sections = append(sections, hpaBehaviorSections(hpa.Spec.Behavior)...)

	summary := component.NewSummary("Configuration", sections...)
	return summary, nil
}

type horizontalPodAutoscalerV2Handler struct {
	horizontalPodAutoScaler *autoscalingv2.HorizontalPodAutoscaler
	configFunc              func(context.Context, *autoscalingv2.HorizontalPodAutoscaler, Options) (*component.Summary, error)
	statusFunc              func(*autoscalingv2.HorizontalPodAutoscaler) (*component.Summary, error)
	metricsFunc             func(context.Context, *autoscalingv2.MetricStatus, Options) (*component.Summary, error)
	conditionsFunc          func(*autoscalingv2.HorizontalPodAutoscaler) (*component.Table, error)
	object                  *Object
}

type horizontalPodAutoscalerV2Object interface {
	Config(ctx context.Context, options Options) error
	Status() error
	Metrics(ctx context.Context, options Options) error
	Conditions() error
}

var _ horizontalPodAutoscalerV2Object = (*horizontalPodAutoscalerV2Handler)(nil)

func newHorizontalPodAutoscalerV2Handler(horizontalPodAutoscaler *autoscalingv2.HorizontalPodAutoscaler, object *Object) (*horizontalPodAutoscalerV2Handler, error) {
	if horizontalPodAutoscaler == nil {
		return nil, errors.New("can't print a nil horizontalpodautoscaler")
	}

	if object == nil {
		return nil, errors.New("can't print horizontalpodautoscaler using a nil object printer")
	}

	hh := &horizontalPodAutoscalerV2Handler{
		horizontalPodAutoScaler: horizontalPodAutoscaler,
		configFunc:              defaultHorizontalPodAutoscalerV2Config,
		statusFunc:              defaultHorizontalPodAutoscalerV2Status,
		metricsFunc:             defaultHorizontalPodAutoscalerV2Metrics,
		conditionsFunc:          defaultHorizontalPodAutoscalerV2Conditions,
		object:                  object,
	}

	return hh, nil
}

func (h *horizontalPodAutoscalerV2Handler) Config(ctx context.Context, options Options) error {
	out, err := h.configFunc(ctx, h.horizontalPodAutoScaler, options)
	if err != nil {
		return err
	}

	h.object.RegisterConfig(out)
	return nil
}

func defaultHorizontalPodAutoscalerV2Config(ctx context.Context, horizontalPodAutoscaler *autoscalingv2.HorizontalPodAutoscaler, options Options) (*component.Summary, error) {
	return NewHorizontalPodAutoscalerV2Configuration(horizontalPodAutoscaler).Create(ctx, options)
}

func (h *horizontalPodAutoscalerV2Handler) Status() error {
	out, err := h.statusFunc(h.horizontalPodAutoScaler)
	if err != nil {
		return err
	}

	h.object.RegisterSummary(out)
	return nil
}

func defaultHorizontalPodAutoscalerV2Status(horizontalPodAutoscaler *autoscalingv2.HorizontalPodAutoscaler) (*component.Summary, error) {
	return createHorizontalPodAutoscalerV2SummaryStatus(horizontalPodAutoscaler)
}

func (h *horizontalPodAutoscalerV2Handler) metrics(ctx context.Context, currentMetrics []autoscalingv2.MetricStatus, options Options) error {
	if h == nil || h.horizontalPodAutoScaler == nil {
		return errors.New("can't display metrics for nil horizontalpodautoscaler")
	}

	for i := range currentMetrics {
		metric := currentMetrics[i]

		if metric.Type == "" {
			continue
		}

		h.object.RegisterItems(ItemDescriptor{
			Width: component.WidthFull,
			Func: func() (component.Component, error) {
				return h.metricsFunc(ctx, &metric, options)
			},
		})
	}

	return nil
}

func (h *horizontalPodAutoscalerV2Handler) Metrics(ctx context.Context, options Options) error {
	if h.horizontalPodAutoScaler == nil {
		return errors.New("can't display metrics for nil horizontalpodautoscaler")
	}

	return h.metrics(ctx, h.horizontalPodAutoScaler.Status.CurrentMetrics, options)
}

func defaultHorizontalPodAutoscalerV2Metrics(ctx context.Context, metricStatus *autoscalingv2.MetricStatus, options Options) (*component.Summary, error) {
	return createHorizontalPodAutoscalerV2MetricsStatusView(metricStatus, options)
}

func (h *horizontalPodAutoscalerV2Handler) Conditions() error {
	if h.horizontalPodAutoScaler == nil {
		return errors.New("can't display conditions for nil horizontalpodautoscaler")
	}

	h.object.RegisterItems(ItemDescriptor{
		Width: component.WidthFull,
		Func: func() (component.Component, error) {
			return h.conditionsFunc(h.horizontalPodAutoScaler)
		},
	})

	return nil
}

func defaultHorizontalPodAutoscalerV2Conditions(horizontalPodAutoscaler *autoscalingv2.HorizontalPodAutoscaler) (*component.Table, error) {
	return createHorizontalPodAutoscalerV2ConditionsView(horizontalPodAutoscaler)
}

func createHorizontalPodAutoscalerV2SummaryStatus(horizontalPodAutoscaler *autoscalingv2.HorizontalPodAutoscaler) (*component.Summary, error) {
	if horizontalPodAutoscaler == nil {
		return nil, errors.New("unable to generate status for a nil horizontalpodautoscaler")
	}

	status := horizontalPodAutoscaler.Status

	summary := component.NewSummary("Status")

	sections := component.SummarySections{}

	sections.AddText("Targets", getCombinedMetricsV2(horizontalPodAutoscaler))

	if status.ObservedGeneration != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Observed Generation",
			Content: component.NewText(fmt.Sprintf("%d", *status.ObservedGeneration)),
		})
	}

	if status.LastScaleTime != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Last Scale Time",
			Content: component.NewTimestamp(status.LastScaleTime.Time),
		})
	}

	sections.AddText("Current Replicas", fmt.Sprintf("%d", status.CurrentReplicas))
	sections.AddText("Desired Replicas", fmt.Sprintf("%d", status.DesiredReplicas))

	if cpu, ok := currentCPUUtilizationV2(status.CurrentMetrics); ok {
		sections = append(sections, component.SummarySection{
			Header:  "Current CPU Utilization Percentage",
			Content: component.NewText(fmt.Sprintf("%d", cpu)),
		})
	}

	summary.Add(sections...)

	return summary, nil
}

func createHorizontalPodAutoscalerV2MetricsStatusView(metricStatus *autoscalingv2.MetricStatus, options Options) (*component.Summary, error) {
	if metricStatus == nil {
		return nil, errors.New("unable to generate metrics from a nil metric status")
	}

	sections := component.SummarySections{}
	summary := component.NewSummary("Metric")

	sections.AddText("Type", fmt.Sprintf("%s", metricStatus.Type))

	if metricStatus.Object != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Name",
			Content: component.NewText(metricStatus.Object.Metric.Name),
		})

		if metricStatus.Object.DescribedObject.Name != "" {
			sections = append(sections, component.SummarySection{
				Header:  "Described Object Name",
				Content: component.NewText(metricStatus.Object.DescribedObject.Name),
			})
			sections = append(sections, component.SummarySection{
				Header:  "Described Object API Version",
				Content: component.NewText(metricStatus.Object.DescribedObject.APIVersion),
			})
			sections = append(sections, component.SummarySection{
				Header:  "Described Object Kind",
				Content: component.NewText(metricStatus.Object.DescribedObject.Kind),
			})
		}

		sections = append(sections, metricValueSections(metricStatus.Object.Current)...)
	}

	if metricStatus.Pods != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Name",
			Content: component.NewText(metricStatus.Pods.Metric.Name),
		})
		sections = append(sections, metricValueSections(metricStatus.Pods.Current)...)
	}

	if metricStatus.Resource != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Name",
			Content: component.NewText(string(metricStatus.Resource.Name)),
		})
		sections = append(sections, metricValueSections(metricStatus.Resource.Current)...)
	}

	if metricStatus.ContainerResource != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Name",
			Content: component.NewText(string(metricStatus.ContainerResource.Name)),
		})
		sections = append(sections, component.SummarySection{
			Header:  "Container",
			Content: component.NewText(metricStatus.ContainerResource.Container),
		})
		sections = append(sections, metricValueSections(metricStatus.ContainerResource.Current)...)
	}

	if metricStatus.External != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Name",
			Content: component.NewText(metricStatus.External.Metric.Name),
		})
		sections = append(sections, metricValueSections(metricStatus.External.Current)...)
	}

	summary.Add(sections...)

	return summary, nil
}

func metricValueSections(current autoscalingv2.MetricValueStatus) []component.SummarySection {
	var sections []component.SummarySection

	if current.AverageUtilization != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Average Utilization",
			Content: component.NewText(fmt.Sprint(*current.AverageUtilization)),
		})
	}

	if current.AverageValue != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Average Value",
			Content: component.NewText(current.AverageValue.String()),
		})
	}

	if current.Value != nil {
		sections = append(sections, component.SummarySection{
			Header:  "Value",
			Content: component.NewText(current.Value.String()),
		})
	}

	return sections
}

func createHorizontalPodAutoscalerV2ConditionsView(horizontalPodAutoscaler *autoscalingv2.HorizontalPodAutoscaler) (*component.Table, error) {
	if horizontalPodAutoscaler == nil {
		return nil, errors.New("unable to generate conditions for a nil horizontalpodautoscaler")
	}

	object, err := runtime.DefaultUnstructuredConverter.ToUnstructured(horizontalPodAutoscaler)
	if err != nil {
		return nil, err
	}

	conditions, err := parseConditions(unstructured.Unstructured{Object: object})
	if err != nil {
		if strings.Contains(err.Error(), "no status found") {
			return createConditionsTable(make([]interface{}, 0), conditionType, hpaConditionColumns), nil
		}
		return nil, err
	}

	return createConditionsTable(conditions, conditionType, hpaConditionColumns), nil
}

func minReplicasV2(horizontalPodAutoscaler *autoscalingv2.HorizontalPodAutoscaler) int32 {
	if horizontalPodAutoscaler == nil || horizontalPodAutoscaler.Spec.MinReplicas == nil {
		return 1
	}

	return *horizontalPodAutoscaler.Spec.MinReplicas
}

func currentCPUUtilizationV2(metrics []autoscalingv2.MetricStatus) (int32, bool) {
	for _, metric := range metrics {
		if metric.Type != autoscalingv2.ResourceMetricSourceType || metric.Resource == nil {
			continue
		}

		if metric.Resource.Name == corev1.ResourceCPU && metric.Resource.Current.AverageUtilization != nil {
			return *metric.Resource.Current.AverageUtilization, true
		}
	}

	return 0, false
}

func getCombinedMetricsV2(horizontalPodAutoscaler *autoscalingv2.HorizontalPodAutoscaler) string {
	if horizontalPodAutoscaler == nil {
		return ""
	}

	var targets = make(map[string]string)
	var currents = make(map[string]string)

	for _, m := range horizontalPodAutoscaler.Spec.Metrics {
		switch m.Type {
		case autoscalingv2.ObjectMetricSourceType:
			if m.Object != nil && m.Object.Metric.Name != "" {
				targets[m.Object.Metric.Name] = metricTargetString(m.Object.Target)
			}
		case autoscalingv2.PodsMetricSourceType:
			if m.Pods != nil && m.Pods.Metric.Name != "" {
				targets[m.Pods.Metric.Name] = metricTargetString(m.Pods.Target)
			}
		case autoscalingv2.ResourceMetricSourceType:
			if m.Resource != nil && m.Resource.Name != "" {
				targets[string(m.Resource.Name)] = metricTargetString(m.Resource.Target)
			}
		case autoscalingv2.ContainerResourceMetricSourceType:
			if m.ContainerResource != nil && m.ContainerResource.Name != "" {
				targets[string(m.ContainerResource.Name)] = metricTargetString(m.ContainerResource.Target)
			}
		case autoscalingv2.ExternalMetricSourceType:
			if m.External != nil && m.External.Metric.Name != "" {
				targets[m.External.Metric.Name] = metricTargetString(m.External.Target)
			}
		}
	}

	for i := range horizontalPodAutoscaler.Status.CurrentMetrics {
		m := horizontalPodAutoscaler.Status.CurrentMetrics[i]

		var name, current string
		switch m.Type {
		case autoscalingv2.ObjectMetricSourceType:
			if m.Object != nil {
				name = m.Object.Metric.Name
				current = metricValueStatusString(m.Object.Current)
			}
		case autoscalingv2.PodsMetricSourceType:
			if m.Pods != nil {
				name = m.Pods.Metric.Name
				current = metricValueStatusString(m.Pods.Current)
			}
		case autoscalingv2.ResourceMetricSourceType:
			if m.Resource != nil {
				name = string(m.Resource.Name)
				current = metricValueStatusString(m.Resource.Current)
			}
		case autoscalingv2.ContainerResourceMetricSourceType:
			if m.ContainerResource != nil {
				name = string(m.ContainerResource.Name)
				current = metricValueStatusString(m.ContainerResource.Current)
			}
		case autoscalingv2.ExternalMetricSourceType:
			if m.External != nil {
				name = m.External.Metric.Name
				current = metricValueStatusString(m.External.Current)
			}
		}

		if name != "" {
			currents[name] = current
		}
	}

	var result []string
	keys := make([]string, 0, len(targets))
	for k := range targets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		if currents[k] == "" {
			currents[k] = "<unknown>"
		}
		result = append(result, currents[k]+"/"+targets[k])
	}

	return strings.Join(result, ", ")
}

func metricTargetString(target autoscalingv2.MetricTarget) string {
	if target.AverageUtilization != nil {
		return fmt.Sprintf("%d%%", *target.AverageUtilization)
	}

	if target.AverageValue != nil {
		return target.AverageValue.String()
	}

	if target.Value != nil {
		return target.Value.String()
	}

	return ""
}

func metricValueStatusString(current autoscalingv2.MetricValueStatus) string {
	if current.AverageUtilization != nil {
		return fmt.Sprintf("%d%%", *current.AverageUtilization)
	}

	if current.AverageValue != nil {
		return current.AverageValue.String()
	}

	if current.Value != nil {
		return current.Value.String()
	}

	return ""
}
