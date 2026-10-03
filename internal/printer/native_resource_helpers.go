/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// formatBoolPtr renders an optional bool as text. A nil value is rendered as an
// empty string so callers can decide how to display an unset field.
func formatBoolPtr(b *bool) string {
	if b == nil {
		return ""
	}
	return fmt.Sprintf("%v", *b)
}

// formatStringPtr renders an optional string as text.
func formatStringPtr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// formatInt32Ptr renders an optional int32 as text.
func formatInt32Ptr(i *int32) string {
	if i == nil {
		return ""
	}
	return fmt.Sprintf("%d", *i)
}

// formatQuantityPtr renders an optional resource quantity as text.
func formatQuantityPtr(q *resource.Quantity) string {
	if q == nil {
		return ""
	}
	return q.String()
}

// formatIntOrString renders an optional IntOrString as text.
func formatIntOrString(v *intstr.IntOrString) string {
	if v == nil {
		return ""
	}
	return v.String()
}

// formatQuantityOrDash renders a resource quantity, falling back to a dash when
// the quantity is unset.
func formatQuantityOrDash(q resource.Quantity) string {
	if q.IsZero() {
		return "-"
	}
	return q.String()
}
