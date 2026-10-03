/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"

	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// ValidatingAdmissionPolicyListHandler is a printFunc that prints validating admission policies.
func ValidatingAdmissionPolicyListHandler(ctx context.Context, list *admissionregistrationv1.ValidatingAdmissionPolicyList, options Options) (component.Component, error) {
	if list == nil {
		return nil, errors.New("validating admission policy list is nil")
	}

	cols := component.NewTableCols("Name", "Failure Policy", "Age")
	ot := NewObjectTable("Validating Admission Policies", "We couldn't find any validating admission policies!", cols, options.DashConfig.ObjectStore())
	ot.EnablePluginStatus(options.DashConfig.PluginManager())
	for _, policy := range list.Items {
		row := component.TableRow{}
		nameLink, err := options.Link.ForObject(&policy, policy.Name)
		if err != nil {
			return nil, err
		}

		row["Name"] = nameLink
		row["Failure Policy"] = component.NewText(failurePolicyString(policy.Spec.FailurePolicy))
		row["Age"] = component.NewTimestamp(policy.CreationTimestamp.Time)

		if err := ot.AddRowForObject(ctx, &policy, row); err != nil {
			return nil, fmt.Errorf("add row for object: %w", err)
		}
	}

	return ot.ToComponent()
}

// ValidatingAdmissionPolicyHandler is a printFunc that prints a single validating admission policy.
func ValidatingAdmissionPolicyHandler(ctx context.Context, policy *admissionregistrationv1.ValidatingAdmissionPolicy, options Options) (component.Component, error) {
	o := NewObject(policy)
	o.EnableEvents()

	config, err := validatingAdmissionPolicyConfig(policy)
	if err != nil {
		return nil, err
	}
	o.RegisterConfig(config)

	return o.ToComponent(ctx, options)
}

func validatingAdmissionPolicyConfig(policy *admissionregistrationv1.ValidatingAdmissionPolicy) (*component.Summary, error) {
	if policy == nil {
		return nil, errors.New("validating admission policy is nil")
	}

	sections := component.SummarySections{}
	sections.AddText("Failure Policy", failurePolicyString(policy.Spec.FailurePolicy))
	sections.AddText("Match Constraints", yesNo(policy.Spec.MatchConstraints != nil))
	sections.AddText("Validations", fmt.Sprintf("%d", len(policy.Spec.Validations)))
	sections.AddText("Audit Annotations", fmt.Sprintf("%d", len(policy.Spec.AuditAnnotations)))
	sections.AddText("Match Conditions", fmt.Sprintf("%d", len(policy.Spec.MatchConditions)))
	sections.AddText("Variables", fmt.Sprintf("%d", len(policy.Spec.Variables)))

	return component.NewSummary("Configuration", sections...), nil
}

func failurePolicyString(policy *admissionregistrationv1.FailurePolicyType) string {
	if policy == nil {
		return ""
	}
	return string(*policy)
}

func yesNo(value bool) string {
	if value {
		return "Yes"
	}
	return "No"
}
