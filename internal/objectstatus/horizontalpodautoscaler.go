/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectstatus

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/vmware-tanzu/octant/internal/link"
	"github.com/vmware-tanzu/octant/pkg/store"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// horizontalPodAutoscaler creates status for an autoscaling/v2 HorizontalPodAutoscaler.
func horizontalPodAutoscaler(_ context.Context, object runtime.Object, _ store.Store, _ link.Interface) (ObjectStatus, error) {
	if object == nil {
		return ObjectStatus{}, errors.Errorf("horizontal pod autoscaler is nil")
	}

	hpa := &autoscalingv2.HorizontalPodAutoscaler{}
	if err := scheme.Scheme.Convert(object, hpa, 0); err != nil {
		return ObjectStatus{}, errors.Wrap(err, "convert object to horizontal pod autoscaler")
	}

	var status ObjectStatus

	if c := findHPACondition(hpa.Status.Conditions, autoscalingv2.AbleToScale); c != nil && c.Status == corev1.ConditionFalse {
		status.SetError()
		status.AddDetailf("Cannot scale: (%s) %s", c.Reason, c.Message)
	}

	if c := findHPACondition(hpa.Status.Conditions, autoscalingv2.ScalingActive); c != nil && c.Status == corev1.ConditionFalse {
		status.SetWarning()
		status.AddDetailf("Scaling is not active: (%s) %s", c.Reason, c.Message)
	}

	if c := findHPACondition(hpa.Status.Conditions, autoscalingv2.ScalingLimited); c != nil && c.Status == corev1.ConditionTrue {
		status.SetWarning()
		status.AddDetailf("Scaling is limited: (%s) %s", c.Reason, c.Message)
	}

	if len(status.Details) == 0 {
		status.AddDetail("Horizontal Pod Autoscaler is OK")
	}

	minimum := "<none>"
	if hpa.Spec.MinReplicas != nil {
		minimum = fmt.Sprintf("%d", *hpa.Spec.MinReplicas)
	}

	current := hpa.Status.CurrentReplicas
	desired := hpa.Status.DesiredReplicas

	status.AddProperty("Target", component.NewText(fmt.Sprintf("%s/%s", hpa.Spec.ScaleTargetRef.Kind, hpa.Spec.ScaleTargetRef.Name)))
	status.AddProperty("Minimum Replicas", component.NewText(minimum))
	status.AddProperty("Maximum Replicas", component.NewText(fmt.Sprintf("%d", hpa.Spec.MaxReplicas)))
	status.AddProperty("Current Replicas", component.NewText(fmt.Sprintf("%d", current)))
	status.AddProperty("Desired Replicas", component.NewText(fmt.Sprintf("%d", desired)))

	return status, nil
}

func findHPACondition(conditions []autoscalingv2.HorizontalPodAutoscalerCondition, conditionType autoscalingv2.HorizontalPodAutoscalerConditionType) *autoscalingv2.HorizontalPodAutoscalerCondition {
	for i := range conditions {
		if conditions[i].Type == conditionType {
			return &conditions[i]
		}
	}
	return nil
}
