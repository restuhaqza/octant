/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectstatus

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/vmware-tanzu/octant/internal/link"
	"github.com/vmware-tanzu/octant/pkg/store"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

// mutatingWebhookConfiguration creates status for an admissionregistration.k8s.io/v1
// MutatingWebhookConfiguration. Webhook configurations have no status subresource,
// so this surfaces the configuration and warns when a webhook can block requests.
func mutatingWebhookConfiguration(_ context.Context, object runtime.Object, _ store.Store, _ link.Interface) (ObjectStatus, error) {
	if object == nil {
		return ObjectStatus{}, errors.Errorf("mutating webhook configuration is nil")
	}

	config := &admissionregistrationv1.MutatingWebhookConfiguration{}
	if err := scheme.Scheme.Convert(object, config, 0); err != nil {
		return ObjectStatus{}, errors.Wrap(err, "convert object to mutating webhook configuration")
	}

	policies := make([]*admissionregistrationv1.FailurePolicyType, 0, len(config.Webhooks))
	for i := range config.Webhooks {
		policies = append(policies, config.Webhooks[i].FailurePolicy)
	}

	return webhookStatus(len(config.Webhooks), countFailPolicies(policies)), nil
}

// validatingWebhookConfiguration creates status for an admissionregistration.k8s.io/v1
// ValidatingWebhookConfiguration.
func validatingWebhookConfiguration(_ context.Context, object runtime.Object, _ store.Store, _ link.Interface) (ObjectStatus, error) {
	if object == nil {
		return ObjectStatus{}, errors.Errorf("validating webhook configuration is nil")
	}

	config := &admissionregistrationv1.ValidatingWebhookConfiguration{}
	if err := scheme.Scheme.Convert(object, config, 0); err != nil {
		return ObjectStatus{}, errors.Wrap(err, "convert object to validating webhook configuration")
	}

	policies := make([]*admissionregistrationv1.FailurePolicyType, 0, len(config.Webhooks))
	for i := range config.Webhooks {
		policies = append(policies, config.Webhooks[i].FailurePolicy)
	}

	return webhookStatus(len(config.Webhooks), countFailPolicies(policies)), nil
}

func webhookStatus(webhooks, failing int) ObjectStatus {
	var status ObjectStatus

	status.AddDetailf("%d webhook(s) configured", webhooks)

	if failing > 0 {
		status.SetWarning()
		status.AddDetailf("%d webhook(s) use failurePolicy=Fail; requests are rejected if the backing service is unreachable", failing)
	}

	status.AddProperty("Webhooks", component.NewText(fmt.Sprintf("%d", webhooks)))
	status.AddProperty("Failure Policy", component.NewText(fmt.Sprintf("%d Fail, %d Ignore", failing, webhooks-failing)))

	return status
}

// countFailPolicies counts webhooks that fail closed. An unset failurePolicy
// defaults to Fail, per the admissionregistration API.
func countFailPolicies(policies []*admissionregistrationv1.FailurePolicyType) int {
	count := 0
	for _, policy := range policies {
		if policy == nil || *policy == admissionregistrationv1.Fail {
			count++
		}
	}
	return count
}
