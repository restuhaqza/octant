/*
Copyright (c) 2019 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/vmware-tanzu/octant/internal/testutil"
	"github.com/vmware-tanzu/octant/pkg/store"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

func Test_HorizontalPodAutoscalerListHandler(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	tpo := newTestPrinterOptions(controller)
	printOptions := tpo.ToOptions()

	objectLabels := map[string]string{
		"foo": "bar",
	}

	now := testutil.Time()
	var minReplicas int32 = 1
	var currentPercentCPU int32 = 5
	var targetPercentCPU int32 = 50

	object := &autoscalingv1.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v1",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "horizontalpodautoscaler",
			Namespace: "default",
			CreationTimestamp: metav1.Time{
				Time: now,
			},
			Labels: objectLabels,
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			MinReplicas:                    &minReplicas,
			MaxReplicas:                    10,
			TargetCPUUtilizationPercentage: &targetPercentCPU,
		},
		Status: autoscalingv1.HorizontalPodAutoscalerStatus{
			CurrentReplicas:                 2,
			CurrentCPUUtilizationPercentage: &currentPercentCPU,
		},
	}

	metrics := `[{"type":"Resource","resource":{"name":"memory","targetAverageUtilization":80}}]`
	object.Annotations = map[string]string{
		"autoscaling.alpha.kubernetes.io/metrics": metrics,
	}

	tpo.PathForObject(object, object.Name, "/path")

	list := &autoscalingv1.HorizontalPodAutoscalerList{
		Items: []autoscalingv1.HorizontalPodAutoscaler{*object},
	}

	ctx := context.Background()
	tpo.pluginManager.EXPECT().ObjectStatus(ctx, object)
	got, err := HorizontalPodAutoscalerListHandler(ctx, list, printOptions)
	require.NoError(t, err)

	cols := component.NewTableCols("Name", "Labels", "Targets", "Minimum Pods", "Maximum Pods", "Replicas", "Age")
	expected := component.NewTable("Horizontal Pod Autoscalers", "We couldn't find any horizontal pod autoscalers", cols)
	expected.Add(component.TableRow{
		"Name": component.NewLink("", "horizontalpodautoscaler", "/path",
			genObjectStatus(component.TextStatusOK, []string{"autoscaling/v1 HorizontalPodAutoscaler is OK"})),
		"Labels":       component.NewLabels(objectLabels),
		"Targets":      component.NewText("5/50%"),
		"Minimum Pods": component.NewText("1"),
		"Maximum Pods": component.NewText("10"),
		"Replicas":     component.NewText("2"),
		"Age":          component.NewTimestamp(now),
		component.GridActionKey: gridActionsFactory([]component.GridAction{
			buildObjectDeleteAction(t, object),
		}),
	})

	component.AssertEqual(t, expected, got)
}

func Test_HorizontalPodAutoscalerConfiguration(t *testing.T) {
	var replicas int32 = 1
	hpa := testutil.CreateHorizontalPodAutoscaler("hpa")
	hpa.Spec.MinReplicas = &replicas
	hpa.Spec.MaxReplicas = 10

	deployment := testutil.CreateDeployment("deployment")
	hpa.Spec.ScaleTargetRef = autoscalingv1.CrossVersionObjectReference{
		Kind:       deployment.Kind,
		APIVersion: deployment.APIVersion,
		Name:       deployment.Name,
	}

	cases := []struct {
		name                    string
		horizontalPodAutoscaler *autoscalingv1.HorizontalPodAutoscaler
		expected                component.Component
		isErr                   bool
	}{
		{
			name:                    "general",
			horizontalPodAutoscaler: hpa,
			expected: component.NewSummary("Configuration", []component.SummarySection{
				{
					Header:  "Scale target",
					Content: component.NewLink("", "deployment", "/deployment"),
				},
				{
					Header:  "Min Replicas",
					Content: component.NewText("1"),
				},
				{
					Header:  "Max Replicas",
					Content: component.NewText("10"),
				},
			}...),
		},
		{
			name:                    "nil horizontalpodautoscaler",
			horizontalPodAutoscaler: nil,
			isErr:                   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			ctx := context.Background()

			tpo := newTestPrinterOptions(controller)
			printOptions := tpo.ToOptions()

			hc := NewHorizontalPodAutoscalerConfiguration(tc.horizontalPodAutoscaler)

			scaleTarget := component.NewLink("", "deployment", "/deployment")
			tpo.link.EXPECT().
				ForGVK("namespace", "apps/v1", "Deployment", "deployment", "deployment").
				Return(scaleTarget, nil).
				AnyTimes()

			if tc.horizontalPodAutoscaler != nil {
				key := store.Key{
					APIVersion: deployment.APIVersion,
					Kind:       deployment.Kind,
					Name:       deployment.Name,
					Namespace:  deployment.Namespace,
				}
				tpo.objectStore.EXPECT().Get(ctx, gomock.Eq(key)).Return(testutil.ToUnstructured(t, deployment), nil)
			}

			summary, err := hc.Create(ctx, printOptions)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			require.NoError(t, err)

			component.AssertEqual(t, tc.expected, summary)
		})
	}
}

func Test_HorizontalPodAutoscalerConfiguration_behaviorAnnotation(t *testing.T) {
	var replicas int32 = 1

	hpa := testutil.CreateHorizontalPodAutoscaler("hpa")
	hpa.Spec.MinReplicas = &replicas
	hpa.Spec.MaxReplicas = 10

	deployment := testutil.CreateDeployment("deployment")
	hpa.Spec.ScaleTargetRef = autoscalingv1.CrossVersionObjectReference{
		Kind:       deployment.Kind,
		APIVersion: deployment.APIVersion,
		Name:       deployment.Name,
	}

	hpa.Annotations = map[string]string{
		"autoscaling.alpha.kubernetes.io/behavior": `{"scaleUp":{"stabilizationWindowSeconds":0,"selectPolicy":"Max","policies":[{"type":"Pods","value":4,"periodSeconds":60}]},"scaleDown":{"stabilizationWindowSeconds":300,"selectPolicy":"Max","policies":[{"type":"Percent","value":100,"periodSeconds":15}]}}`,
	}

	controller := gomock.NewController(t)
	defer controller.Finish()

	ctx := context.Background()

	tpo := newTestPrinterOptions(controller)
	printOptions := tpo.ToOptions()

	scaleTarget := component.NewLink("", "deployment", "/deployment")
	tpo.link.EXPECT().
		ForGVK("namespace", "apps/v1", "Deployment", "deployment", "deployment").
		Return(scaleTarget, nil).
		AnyTimes()

	key := store.Key{
		APIVersion: deployment.APIVersion,
		Kind:       deployment.Kind,
		Name:       deployment.Name,
		Namespace:  deployment.Namespace,
	}
	tpo.objectStore.EXPECT().Get(ctx, gomock.Eq(key)).Return(testutil.ToUnstructured(t, deployment), nil)

	hc := NewHorizontalPodAutoscalerConfiguration(hpa)

	summary, err := hc.Create(ctx, printOptions)
	require.NoError(t, err)

	expected := component.NewSummary("Configuration", []component.SummarySection{
		{
			Header:  "Scale target",
			Content: component.NewLink("", "deployment", "/deployment"),
		},
		{
			Header:  "Min Replicas",
			Content: component.NewText("1"),
		},
		{
			Header:  "Max Replicas",
			Content: component.NewText("10"),
		},
		{
			Header: "Scale Up",
			Content: component.NewTableWithRows(
				"",
				"There are no scale up policies!",
				component.NewTableCols("Stabilization Window", "Select Policies", "Policies"),
				[]component.TableRow{
					{
						"Stabilization Window": component.NewText("0 seconds"),
						"Select Policies":      component.NewText("Max"),
						"Policies":             component.NewText("4 Pods / 60 seconds"),
					},
				},
			),
		},
		{
			Header: "Scale Down",
			Content: component.NewTableWithRows(
				"",
				"There are no scale down policies!",
				component.NewTableCols("Stabilization Window", "Select Policies", "Policies"),
				[]component.TableRow{
					{
						"Stabilization Window": component.NewText("300 seconds"),
						"Select Policies":      component.NewText("Max"),
						"Policies":             component.NewText("100 Percent / 15 seconds"),
					},
				},
			),
		},
	}...)

	component.AssertEqual(t, expected, summary)
}

func Test_createHorizontalPodAutoscalerSummaryStatus(t *testing.T) {
	var observedGeneration int64 = 1
	var currentCPU int32 = 3
	var targetCPUPercent int32 = 80
	now := testutil.Time()

	hpa := testutil.CreateHorizontalPodAutoscaler("hpa")
	hpa.Status.ObservedGeneration = &observedGeneration
	hpa.Status.LastScaleTime = &metav1.Time{Time: now}
	hpa.Status.CurrentReplicas = 2
	hpa.Status.DesiredReplicas = 7
	hpa.Status.CurrentCPUUtilizationPercentage = &currentCPU
	hpa.Spec.TargetCPUUtilizationPercentage = &targetCPUPercent

	cases := []struct {
		name                    string
		horizontalPodAutoscaler *autoscalingv1.HorizontalPodAutoscaler
		expected                *component.Summary
		isErr                   bool
	}{
		{
			name:                    "in general",
			horizontalPodAutoscaler: hpa,
			expected: component.NewSummary("Status", []component.SummarySection{
				{
					Header:  "Targets",
					Content: component.NewText("3/80%"),
				},
				{
					Header:  "Observed Generation",
					Content: component.NewText("1"),
				},
				{
					Header:  "Last Scale Time",
					Content: component.NewTimestamp(now),
				},
				{
					Header:  "Current Replicas",
					Content: component.NewText("2"),
				},
				{
					Header:  "Desired Replicas",
					Content: component.NewText("7"),
				},
				{
					Header:  "Current CPU Utilization Percentage",
					Content: component.NewText("3"),
				},
			}...),
		},
		{
			name:                    "nil horizontalpodautoscaler",
			horizontalPodAutoscaler: nil,
			isErr:                   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			summary, err := createHorizontalPodAutoscalerSummaryStatus(tc.horizontalPodAutoscaler)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			component.AssertEqual(t, tc.expected, summary)
		})
	}
}

func Test_HorizontalPodAutoscalerMetrics(t *testing.T) {
	var averageUtilization int32 = 20
	metricResource := autoscalingv1.MetricStatus{
		Type: autoscalingv1.ResourceMetricSourceType,
		Resource: &autoscalingv1.ResourceMetricStatus{
			Name:                      "cpu",
			CurrentAverageUtilization: &averageUtilization,
		},
	}

	metricPods := autoscalingv1.MetricStatus{
		Type: autoscalingv1.PodsMetricSourceType,
		Pods: &autoscalingv1.PodsMetricStatus{
			MetricName:          "packets-per-second",
			CurrentAverageValue: *resource.NewMilliQuantity(1000, resource.DecimalSI),
		},
	}

	cases := []struct {
		name         string
		metricStatus *autoscalingv1.MetricStatus
		expected     *component.Summary
		isErr        bool
	}{
		{
			name:         "resource type",
			metricStatus: &metricResource,
			expected: component.NewSummary("Metric", []component.SummarySection{
				{
					Header:  "Type",
					Content: component.NewText("Resource"),
				},
				{
					Header:  "Name",
					Content: component.NewText("cpu"),
				},
				{
					Header:  "Average Utilization",
					Content: component.NewText("20"),
				},
				{
					Header:  "Average Value",
					Content: component.NewText("0"),
				},
			}...),
		},
		{
			name:         "pods type",
			metricStatus: &metricPods,
			expected: component.NewSummary("Metric", []component.SummarySection{
				{
					Header:  "Type",
					Content: component.NewText("Pods"),
				},
				{
					Header:  "Name",
					Content: component.NewText("packets-per-second"),
				},
				{
					Header:  "Average Utilization",
					Content: component.NewText("1"),
				},
			}...),
		},
		{
			name:         "nil metric",
			metricStatus: nil,
			isErr:        true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			tpo := newTestPrinterOptions(controller)
			printOptions := tpo.ToOptions()

			summary, err := createHorizontalPodAutoscalerMetricsStatusView(tc.metricStatus, printOptions)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			component.AssertEqual(t, tc.expected, summary)
		})
	}
}

func Test_createHorizontalPodAutoscalerConditionsView(t *testing.T) {
	now := metav1.Time{Time: testutil.Time()}

	condition := `[{"type":"AbleToScale","status":"True","reason":"reason","message":"message","lastTransitionTime":"2019-01-11T12:57:10Z"}]`

	horizontalPodAutoscaler := testutil.CreateHorizontalPodAutoscaler("hpa")
	horizontalPodAutoscaler.Annotations = map[string]string{
		"autoscaling.alpha.kubernetes.io/conditions": condition,
	}

	got, err := createHorizontalPodAutoscalerConditionsView(horizontalPodAutoscaler)
	require.NoError(t, err)

	cols := component.NewTableCols("Type", "Reason", "Status", "Message", "Last Transition")
	expected := component.NewTable("Conditions", "There are no conditions!", cols)
	expected.Sort("Type")
	expected.Add([]component.TableRow{
		{
			"Type":            component.NewText("AbleToScale"),
			"Reason":          component.NewText("reason"),
			"Status":          component.NewText("True"),
			"Message":         component.NewText("message"),
			"Last Transition": component.NewTimestamp(now.Time),
		},
	}...)

	component.AssertEqual(t, expected, got)
}

func Test_getCombinedMetrics(t *testing.T) {
	hpa := testutil.CreateHorizontalPodAutoscaler("hpa")
	cases := []struct {
		name                    string
		horizontalPodAutoscaler autoscalingv1.HorizontalPodAutoscaler
		annotations             map[string]string
		expected                string
	}{
		{
			name: "Object Metric Value",
			annotations: map[string]string{
				"autoscaling.alpha.kubernetes.io/metrics":         `[{"type":"Object","object":{"metricName":"requests-per-second","targetValue":42}}]`,
				"autoscaling.alpha.kubernetes.io/current-metrics": `[{"type":"Object","object":{"metricName":"requests-per-second","currentValue":1}}]`,
			},
			horizontalPodAutoscaler: *hpa,
			expected:                "1/42",
		},
		{
			name: "Pods Metric Average Value",
			annotations: map[string]string{
				"autoscaling.alpha.kubernetes.io/metrics":         `[{"type":"Pods","pods":{"metricName":"packets-per-second","targetAverageValue":"350"}}]`,
				"autoscaling.alpha.kubernetes.io/current-metrics": `[{"type":"Pods","pods":{"metricName":"packets-per-second","currentAverageValue":"1000"}}]`,
			},
			horizontalPodAutoscaler: *hpa,
			expected:                "1k/350",
		},
		{
			name: "Resource Metric Average Value",
			annotations: map[string]string{
				"autoscaling.alpha.kubernetes.io/metrics":         `[{"type":"Resource","resource":{"name":"memory","targetAverageValue":9000}}]`,
				"autoscaling.alpha.kubernetes.io/current-metrics": `[{"type":"Resource","resource":{"name":"memory","currentAverageValue":"9001"}}]`,
			},
			horizontalPodAutoscaler: *hpa,
			expected:                "9001/9k",
		},
		{
			name: "Resource Metric Average Utilization",
			annotations: map[string]string{
				"autoscaling.alpha.kubernetes.io/metrics":         `[{"type":"Resource","resource":{"name":"memory","targetAverageUtilization":9000}}]`,
				"autoscaling.alpha.kubernetes.io/current-metrics": `[{"type":"Resource","resource":{"name":"memory","currentAverageUtilization":9001}}]`,
			},
			horizontalPodAutoscaler: *hpa,
			expected:                "",
		},
		{
			name: "External Metric Average Value",
			annotations: map[string]string{
				"autoscaling.alpha.kubernetes.io/metrics":         `[{"type":"External","external":{"metricName":"queue_messages_ready","targetAverageValue":"0"}}]`,
				"autoscaling.alpha.kubernetes.io/current-metrics": `[{"type":"External","external":{"metricName":"queue_messages_ready","currentAverageValue":"3"}}]`,
			},
			horizontalPodAutoscaler: *hpa,
			expected:                "3/0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.horizontalPodAutoscaler.Annotations = tc.annotations

			horizontalPodAutoscalerMetrics, horizontalPodAutoscalerCurrentMetrics, err := parseAnnotations(tc.horizontalPodAutoscaler)
			require.NoError(t, err)

			result, err := getCombinedMetrics(tc.horizontalPodAutoscaler, horizontalPodAutoscalerMetrics, horizontalPodAutoscalerCurrentMetrics)
			require.NoError(t, err)

			assert.Equal(t, tc.expected, result)
		})
	}
}

func Test_HorizontalPodAutoscalerV2ListHandler(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	tpo := newTestPrinterOptions(controller)
	printOptions := tpo.ToOptions()

	objectLabels := map[string]string{
		"foo": "bar",
	}

	now := testutil.Time()
	var minReplicas int32 = 1
	var currentPercentCPU int32 = 5
	var targetPercentCPU int32 = 50

	object := &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "horizontalpodautoscaler",
			Namespace: "default",
			CreationTimestamp: metav1.Time{
				Time: now,
			},
			Labels: objectLabels,
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			MinReplicas: &minReplicas,
			MaxReplicas: 10,
			Metrics: []autoscalingv2.MetricSpec{
				{
					Type: autoscalingv2.ResourceMetricSourceType,
					Resource: &autoscalingv2.ResourceMetricSource{
						Name: corev1.ResourceCPU,
						Target: autoscalingv2.MetricTarget{
							Type:               autoscalingv2.UtilizationMetricType,
							AverageUtilization: &targetPercentCPU,
						},
					},
				},
			},
		},
		Status: autoscalingv2.HorizontalPodAutoscalerStatus{
			CurrentReplicas: 2,
			CurrentMetrics: []autoscalingv2.MetricStatus{
				{
					Type: autoscalingv2.ResourceMetricSourceType,
					Resource: &autoscalingv2.ResourceMetricStatus{
						Name: corev1.ResourceCPU,
						Current: autoscalingv2.MetricValueStatus{
							AverageUtilization: &currentPercentCPU,
						},
					},
				},
			},
		},
	}

	tpo.PathForObject(object, object.Name, "/path")

	list := &autoscalingv2.HorizontalPodAutoscalerList{
		Items: []autoscalingv2.HorizontalPodAutoscaler{*object},
	}

	ctx := context.Background()
	tpo.pluginManager.EXPECT().ObjectStatus(ctx, object)
	got, err := HorizontalPodAutoscalerV2ListHandler(ctx, list, printOptions)
	require.NoError(t, err)

	cols := component.NewTableCols("Name", "Labels", "Targets", "Minimum Pods", "Maximum Pods", "Replicas", "Age")
	expected := component.NewTable("Horizontal Pod Autoscalers", "We couldn't find any horizontal pod autoscalers", cols)
	expected.Add(component.TableRow{
		"Name": component.NewLink("", "horizontalpodautoscaler", "/path",
			genObjectStatus(component.TextStatusOK, []string{"autoscaling/v2 HorizontalPodAutoscaler is OK"})),
		"Labels":       component.NewLabels(objectLabels),
		"Targets":      component.NewText("5%/50%"),
		"Minimum Pods": component.NewText("1"),
		"Maximum Pods": component.NewText("10"),
		"Replicas":     component.NewText("2"),
		"Age":          component.NewTimestamp(now),
		component.GridActionKey: gridActionsFactory([]component.GridAction{
			buildObjectDeleteAction(t, object),
		}),
	})

	component.AssertEqual(t, expected, got)
}

func Test_HorizontalPodAutoscalerV2Configuration(t *testing.T) {
	var replicas int32 = 1
	var stabilizationUp int32 = 0
	var stabilizationDown int32 = 300
	maxSelect := autoscalingv2.MaxChangePolicySelect

	deployment := testutil.CreateDeployment("deployment")

	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hpa",
			Namespace: "namespace",
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			MinReplicas: &replicas,
			MaxReplicas: 10,
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				Kind:       deployment.Kind,
				APIVersion: deployment.APIVersion,
				Name:       deployment.Name,
			},
			Behavior: &autoscalingv2.HorizontalPodAutoscalerBehavior{
				ScaleUp: &autoscalingv2.HPAScalingRules{
					StabilizationWindowSeconds: &stabilizationUp,
					SelectPolicy:               &maxSelect,
					Policies: []autoscalingv2.HPAScalingPolicy{
						{Type: autoscalingv2.PodsScalingPolicy, Value: 4, PeriodSeconds: 60},
					},
				},
				ScaleDown: &autoscalingv2.HPAScalingRules{
					StabilizationWindowSeconds: &stabilizationDown,
					SelectPolicy:               &maxSelect,
					Policies: []autoscalingv2.HPAScalingPolicy{
						{Type: autoscalingv2.PercentScalingPolicy, Value: 100, PeriodSeconds: 15},
					},
				},
			},
		},
	}

	cases := []struct {
		name                    string
		horizontalPodAutoscaler *autoscalingv2.HorizontalPodAutoscaler
		expected                component.Component
		isErr                   bool
	}{
		{
			name:                    "general",
			horizontalPodAutoscaler: hpa,
			expected: component.NewSummary("Configuration", []component.SummarySection{
				{
					Header:  "Scale target",
					Content: component.NewLink("", "deployment", "/deployment"),
				},
				{
					Header:  "Min Replicas",
					Content: component.NewText("1"),
				},
				{
					Header:  "Max Replicas",
					Content: component.NewText("10"),
				},
				{
					Header: "Scale Up",
					Content: component.NewTableWithRows(
						"",
						"There are no scale up policies!",
						component.NewTableCols("Stabilization Window", "Select Policies", "Policies"),
						[]component.TableRow{
							{
								"Stabilization Window": component.NewText("0 seconds"),
								"Select Policies":      component.NewText("Max"),
								"Policies":             component.NewText("4 Pods / 60 seconds"),
							},
						},
					),
				},
				{
					Header: "Scale Down",
					Content: component.NewTableWithRows(
						"",
						"There are no scale down policies!",
						component.NewTableCols("Stabilization Window", "Select Policies", "Policies"),
						[]component.TableRow{
							{
								"Stabilization Window": component.NewText("300 seconds"),
								"Select Policies":      component.NewText("Max"),
								"Policies":             component.NewText("100 Percent / 15 seconds"),
							},
						},
					),
				},
			}...),
		},
		{
			name:                    "nil horizontalpodautoscaler",
			horizontalPodAutoscaler: nil,
			isErr:                   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			ctx := context.Background()

			tpo := newTestPrinterOptions(controller)
			printOptions := tpo.ToOptions()

			hc := NewHorizontalPodAutoscalerV2Configuration(tc.horizontalPodAutoscaler)

			scaleTarget := component.NewLink("", "deployment", "/deployment")
			tpo.link.EXPECT().
				ForGVK("namespace", "apps/v1", "Deployment", "deployment", "deployment").
				Return(scaleTarget, nil).
				AnyTimes()

			if tc.horizontalPodAutoscaler != nil {
				key := store.Key{
					APIVersion: deployment.APIVersion,
					Kind:       deployment.Kind,
					Name:       deployment.Name,
					Namespace:  deployment.Namespace,
				}
				tpo.objectStore.EXPECT().Get(ctx, gomock.Eq(key)).Return(testutil.ToUnstructured(t, deployment), nil)
			}

			summary, err := hc.Create(ctx, printOptions)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			component.AssertEqual(t, tc.expected, summary)
		})
	}
}

func Test_createHorizontalPodAutoscalerV2SummaryStatus(t *testing.T) {
	var observedGeneration int64 = 1
	var currentCPU int32 = 3
	var targetCPUPercent int32 = 80
	now := testutil.Time()

	hpa := &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hpa",
			Namespace: "namespace",
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			MaxReplicas: 10,
			Metrics: []autoscalingv2.MetricSpec{
				{
					Type: autoscalingv2.ResourceMetricSourceType,
					Resource: &autoscalingv2.ResourceMetricSource{
						Name: corev1.ResourceCPU,
						Target: autoscalingv2.MetricTarget{
							Type:               autoscalingv2.UtilizationMetricType,
							AverageUtilization: &targetCPUPercent,
						},
					},
				},
			},
		},
		Status: autoscalingv2.HorizontalPodAutoscalerStatus{
			ObservedGeneration: &observedGeneration,
			LastScaleTime:      &metav1.Time{Time: now},
			CurrentReplicas:    2,
			DesiredReplicas:    7,
			CurrentMetrics: []autoscalingv2.MetricStatus{
				{
					Type: autoscalingv2.ResourceMetricSourceType,
					Resource: &autoscalingv2.ResourceMetricStatus{
						Name: corev1.ResourceCPU,
						Current: autoscalingv2.MetricValueStatus{
							AverageUtilization: &currentCPU,
						},
					},
				},
			},
		},
	}

	cases := []struct {
		name                    string
		horizontalPodAutoscaler *autoscalingv2.HorizontalPodAutoscaler
		expected                *component.Summary
		isErr                   bool
	}{
		{
			name:                    "in general",
			horizontalPodAutoscaler: hpa,
			expected: component.NewSummary("Status", []component.SummarySection{
				{
					Header:  "Targets",
					Content: component.NewText("3%/80%"),
				},
				{
					Header:  "Observed Generation",
					Content: component.NewText("1"),
				},
				{
					Header:  "Last Scale Time",
					Content: component.NewTimestamp(now),
				},
				{
					Header:  "Current Replicas",
					Content: component.NewText("2"),
				},
				{
					Header:  "Desired Replicas",
					Content: component.NewText("7"),
				},
				{
					Header:  "Current CPU Utilization Percentage",
					Content: component.NewText("3"),
				},
			}...),
		},
		{
			name:                    "nil horizontalpodautoscaler",
			horizontalPodAutoscaler: nil,
			isErr:                   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			summary, err := createHorizontalPodAutoscalerV2SummaryStatus(tc.horizontalPodAutoscaler)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			component.AssertEqual(t, tc.expected, summary)
		})
	}
}

func Test_HorizontalPodAutoscalerV2Metrics(t *testing.T) {
	var averageUtilization int32 = 20
	metricResource := autoscalingv2.MetricStatus{
		Type: autoscalingv2.ResourceMetricSourceType,
		Resource: &autoscalingv2.ResourceMetricStatus{
			Name: corev1.ResourceCPU,
			Current: autoscalingv2.MetricValueStatus{
				AverageUtilization: &averageUtilization,
				AverageValue:       resource.NewMilliQuantity(0, resource.DecimalSI),
			},
		},
	}

	metricPods := autoscalingv2.MetricStatus{
		Type: autoscalingv2.PodsMetricSourceType,
		Pods: &autoscalingv2.PodsMetricStatus{
			Metric: autoscalingv2.MetricIdentifier{
				Name: "packets-per-second",
			},
			Current: autoscalingv2.MetricValueStatus{
				AverageValue: resource.NewMilliQuantity(1000, resource.DecimalSI),
			},
		},
	}

	cases := []struct {
		name         string
		metricStatus *autoscalingv2.MetricStatus
		expected     *component.Summary
		isErr        bool
	}{
		{
			name:         "resource type",
			metricStatus: &metricResource,
			expected: component.NewSummary("Metric", []component.SummarySection{
				{
					Header:  "Type",
					Content: component.NewText("Resource"),
				},
				{
					Header:  "Name",
					Content: component.NewText("cpu"),
				},
				{
					Header:  "Average Utilization",
					Content: component.NewText("20"),
				},
				{
					Header:  "Average Value",
					Content: component.NewText("0"),
				},
			}...),
		},
		{
			name:         "pods type",
			metricStatus: &metricPods,
			expected: component.NewSummary("Metric", []component.SummarySection{
				{
					Header:  "Type",
					Content: component.NewText("Pods"),
				},
				{
					Header:  "Name",
					Content: component.NewText("packets-per-second"),
				},
				{
					Header:  "Average Value",
					Content: component.NewText("1"),
				},
			}...),
		},
		{
			name:         "nil metric",
			metricStatus: nil,
			isErr:        true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			controller := gomock.NewController(t)
			defer controller.Finish()

			tpo := newTestPrinterOptions(controller)
			printOptions := tpo.ToOptions()

			summary, err := createHorizontalPodAutoscalerV2MetricsStatusView(tc.metricStatus, printOptions)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			component.AssertEqual(t, tc.expected, summary)
		})
	}
}

func Test_createHorizontalPodAutoscalerV2ConditionsView(t *testing.T) {
	now := metav1.Time{Time: testutil.Time()}

	horizontalPodAutoscaler := &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hpa",
			Namespace: "namespace",
		},
		Status: autoscalingv2.HorizontalPodAutoscalerStatus{
			Conditions: []autoscalingv2.HorizontalPodAutoscalerCondition{
				{
					Type:               autoscalingv2.AbleToScale,
					Status:             corev1.ConditionTrue,
					Reason:             "reason",
					Message:            "message",
					LastTransitionTime: now,
				},
			},
		},
	}

	got, err := createHorizontalPodAutoscalerV2ConditionsView(horizontalPodAutoscaler)
	require.NoError(t, err)

	cols := component.NewTableCols("Type", "Reason", "Status", "Message", "Last Transition")
	expected := component.NewTable("Conditions", "There are no conditions!", cols)
	expected.Sort("Type")
	expected.Add([]component.TableRow{
		{
			"Type":            component.NewText("AbleToScale"),
			"Reason":          component.NewText("reason"),
			"Status":          component.NewText("True"),
			"Message":         component.NewText("message"),
			"Last Transition": component.NewTimestamp(now.Time),
		},
	}...)

	component.AssertEqual(t, expected, got)
}

func Test_getCombinedMetricsV2(t *testing.T) {
	cases := []struct {
		name     string
		hpa      *autoscalingv2.HorizontalPodAutoscaler
		expected string
	}{
		{
			name: "resource utilization",
			hpa: autoscalingv2HPAWithMetrics(
				[]autoscalingv2.MetricSpec{
					{
						Type: autoscalingv2.ResourceMetricSourceType,
						Resource: &autoscalingv2.ResourceMetricSource{
							Name: corev1.ResourceCPU,
							Target: autoscalingv2.MetricTarget{
								Type:               autoscalingv2.UtilizationMetricType,
								AverageUtilization: int32Ptr(50),
							},
						},
					},
				},
				[]autoscalingv2.MetricStatus{
					{
						Type: autoscalingv2.ResourceMetricSourceType,
						Resource: &autoscalingv2.ResourceMetricStatus{
							Name: corev1.ResourceCPU,
							Current: autoscalingv2.MetricValueStatus{
								AverageUtilization: int32Ptr(5),
							},
						},
					},
				},
			),
			expected: "5%/50%",
		},
		{
			name: "resource average value",
			hpa: autoscalingv2HPAWithMetrics(
				[]autoscalingv2.MetricSpec{
					{
						Type: autoscalingv2.ResourceMetricSourceType,
						Resource: &autoscalingv2.ResourceMetricSource{
							Name: corev1.ResourceMemory,
							Target: autoscalingv2.MetricTarget{
								Type:         autoscalingv2.AverageValueMetricType,
								AverageValue: resource.NewQuantity(9000, resource.DecimalSI),
							},
						},
					},
				},
				[]autoscalingv2.MetricStatus{
					{
						Type: autoscalingv2.ResourceMetricSourceType,
						Resource: &autoscalingv2.ResourceMetricStatus{
							Name: corev1.ResourceMemory,
							Current: autoscalingv2.MetricValueStatus{
								AverageValue: resource.NewQuantity(9001, resource.DecimalSI),
							},
						},
					},
				},
			),
			expected: "9001/9k",
		},
		{
			name: "external value without current",
			hpa: autoscalingv2HPAWithMetrics(
				[]autoscalingv2.MetricSpec{
					{
						Type: autoscalingv2.ExternalMetricSourceType,
						External: &autoscalingv2.ExternalMetricSource{
							Metric: autoscalingv2.MetricIdentifier{Name: "queue_messages_ready"},
							Target: autoscalingv2.MetricTarget{
								Type:  autoscalingv2.ValueMetricType,
								Value: resource.NewQuantity(0, resource.DecimalSI),
							},
						},
					},
				},
				nil,
			),
			expected: "<unknown>/0",
		},
		{
			name:     "nil hpa",
			hpa:      nil,
			expected: "",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, getCombinedMetricsV2(tc.hpa))
		})
	}
}

func autoscalingv2HPAWithMetrics(metrics []autoscalingv2.MetricSpec, status []autoscalingv2.MetricStatus) *autoscalingv2.HorizontalPodAutoscaler {
	return &autoscalingv2.HorizontalPodAutoscaler{
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			Metrics: metrics,
		},
		Status: autoscalingv2.HorizontalPodAutoscalerStatus{
			CurrentMetrics: status,
		},
	}
}

func int32Ptr(v int32) *int32 {
	return &v
}
