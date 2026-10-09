/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package objectstatus

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"

	"github.com/vmware-tanzu/octant/internal/testutil"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

func Test_mutatingWebhookConfiguration(t *testing.T) {
	fail := admissionregistrationv1.Fail
	ignore := admissionregistrationv1.Ignore

	cases := []struct {
		name     string
		object   runtime.Object
		expected ObjectStatus
		isErr    bool
	}{
		{
			name: "only ignore webhooks",
			object: func() runtime.Object {
				config := testutil.CreateMutatingWebhookConfiguration("webhook")
				config.Webhooks = []admissionregistrationv1.MutatingWebhook{
					{Name: "ignore.example.com", FailurePolicy: &ignore},
				}
				return config
			}(),
			expected: ObjectStatus{
				Details: []component.Component{component.NewText("1 webhook(s) configured")},
				Properties: []component.Property{
					{Label: "Webhooks", Value: component.NewText("1")},
					{Label: "Failure Policy", Value: component.NewText("0 Fail, 1 Ignore")},
				},
			},
		},
		{
			name: "failing webhooks",
			object: func() runtime.Object {
				config := testutil.CreateMutatingWebhookConfiguration("webhook")
				config.Webhooks = []admissionregistrationv1.MutatingWebhook{
					{Name: "fail.example.com", FailurePolicy: &fail},
					{Name: "ignore.example.com", FailurePolicy: &ignore},
				}
				return config
			}(),
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusWarning,
				Details: []component.Component{
					component.NewText("2 webhook(s) configured"),
					component.NewText("1 webhook(s) use failurePolicy=Fail; requests are rejected if the backing service is unreachable"),
				},
				Properties: []component.Property{
					{Label: "Webhooks", Value: component.NewText("2")},
					{Label: "Failure Policy", Value: component.NewText("1 Fail, 1 Ignore")},
				},
			},
		},
		{
			name: "unset policy defaults to fail",
			object: func() runtime.Object {
				config := testutil.CreateMutatingWebhookConfiguration("webhook")
				config.Webhooks = []admissionregistrationv1.MutatingWebhook{
					{Name: "default.example.com"},
				}
				return config
			}(),
			expected: ObjectStatus{
				NodeStatus: component.NodeStatusWarning,
				Details: []component.Component{
					component.NewText("1 webhook(s) configured"),
					component.NewText("1 webhook(s) use failurePolicy=Fail; requests are rejected if the backing service is unreachable"),
				},
				Properties: []component.Property{
					{Label: "Webhooks", Value: component.NewText("1")},
					{Label: "Failure Policy", Value: component.NewText("1 Fail, 0 Ignore")},
				},
			},
		},
		{
			name:   "nil object",
			object: nil,
			isErr:  true,
		},
		{
			name:   "not a mutating webhook configuration",
			object: &unstructured.Unstructured{},
			isErr:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := mutatingWebhookConfiguration(context.Background(), tc.object, nil, nil)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func Test_validatingWebhookConfiguration(t *testing.T) {
	cases := []struct {
		name     string
		object   runtime.Object
		expected ObjectStatus
		isErr    bool
	}{
		{
			name: "webhooks configured",
			object: func() runtime.Object {
				config := testutil.CreateValidatingWebhookConfiguration("webhook")
				ignore := admissionregistrationv1.Ignore
				config.Webhooks = []admissionregistrationv1.ValidatingWebhook{
					{Name: "validate.example.com", FailurePolicy: &ignore},
				}
				return config
			}(),
			expected: ObjectStatus{
				Details: []component.Component{component.NewText("1 webhook(s) configured")},
				Properties: []component.Property{
					{Label: "Webhooks", Value: component.NewText("1")},
					{Label: "Failure Policy", Value: component.NewText("0 Fail, 1 Ignore")},
				},
			},
		},
		{
			name:   "nil object",
			object: nil,
			isErr:  true,
		},
		{
			name:   "not a validating webhook configuration",
			object: &unstructured.Unstructured{},
			isErr:  true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := validatingWebhookConfiguration(context.Background(), tc.object, nil, nil)
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expected, got)
		})
	}
}

func init() {
	_ = admissionregistrationv1.AddToScheme(scheme.Scheme)
}
