/*
Copyright (c) 2026 the Octant contributors. All Rights Reserved.
SPDX-License-Identifier: Apache-2.0
*/

package printer

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	coordinationv1 "k8s.io/api/coordination/v1"
	corev1 "k8s.io/api/core/v1"
	flowcontrolv1 "k8s.io/api/flowcontrol/v1"
	networkingv1 "k8s.io/api/networking/v1"
	nodev1 "k8s.io/api/node/v1"
	policyv1 "k8s.io/api/policy/v1"
	schedulingv1 "k8s.io/api/scheduling/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/vmware-tanzu/octant/internal/testutil"
	"github.com/vmware-tanzu/octant/pkg/view/component"
)

type nativeListCase struct {
	title       string
	placeholder string
	cols        []component.TableCol
	object      runtime.Object
	row         component.TableRow
	status      component.TextStatus
	messages    []string
	call        func(ctx context.Context, options Options) (component.Component, error)
}

// runNativeListTest runs a list handler for a single object and asserts the
// produced table row, including the object status and delete action.
func runNativeListTest(t *testing.T, tc nativeListCase) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	tpo := newTestPrinterOptions(controller)
	printOptions := tpo.ToOptions()
	ctx := context.Background()

	accessor, err := meta.Accessor(tc.object)
	require.NoError(t, err)
	name := accessor.GetName()

	tpo.PathForObject(tc.object, name, "/"+name)
	tpo.pluginManager.EXPECT().ObjectStatus(ctx, tc.object)

	got, err := tc.call(ctx, printOptions)
	require.NoError(t, err)

	row := tc.row
	row["Name"] = component.NewLink("", name, "/"+name, genObjectStatus(tc.status, tc.messages))
	row[component.GridActionKey] = gridActionsFactory([]component.GridAction{
		buildObjectDeleteAction(t, tc.object),
	})

	expected := component.NewTableWithRows(tc.title, tc.placeholder, tc.cols, []component.TableRow{row})
	component.AssertEqual(t, expected, got)
}

func Test_PodDisruptionBudgetListHandler(t *testing.T) {
	now := testutil.Time()
	object := testutil.CreatePodDisruptionBudget("pdb")
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &policyv1.PodDisruptionBudgetList{Items: []policyv1.PodDisruptionBudget{*object}}

	cols := component.NewTableCols("Name", "Min Available", "Max Unavailable", "Allowed Disruptions", "Age")

	runNativeListTest(t, nativeListCase{
		title:       "Pod Disruption Budgets",
		placeholder: "We couldn't find any pod disruption budgets!",
		cols:        cols,
		object:      object,
		row: component.TableRow{
			"Min Available":       component.NewText(""),
			"Max Unavailable":     component.NewText("1"),
			"Allowed Disruptions": component.NewText("0"),
			"Age":                 component.NewTimestamp(now),
		},
		status:   component.TextStatusOK,
		messages: []string{"0 disruptions allowed"},
		call: func(ctx context.Context, options Options) (component.Component, error) {
			return PodDisruptionBudgetListHandler(ctx, list, options)
		},
	})

	_, err := PodDisruptionBudgetListHandler(context.Background(), nil, Options{})
	require.Error(t, err)
}

func Test_ResourceQuotaListHandler(t *testing.T) {
	now := testutil.Time()
	object := testutil.CreateResourceQuota("quota")
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &corev1.ResourceQuotaList{Items: []corev1.ResourceQuota{*object}}

	cols := component.NewTableCols("Name", "Scopes", "Age")

	runNativeListTest(t, nativeListCase{
		title:       "Resource Quotas",
		placeholder: "We couldn't find any resource quotas!",
		cols:        cols,
		object:      object,
		row: component.TableRow{
			"Scopes": component.NewText("-"),
			"Age":    component.NewTimestamp(now),
		},
		status:   component.TextStatusOK,
		messages: []string{"v1 ResourceQuota is OK"},
		call: func(ctx context.Context, options Options) (component.Component, error) {
			return ResourceQuotaListHandler(ctx, list, options)
		},
	})

	_, err := ResourceQuotaListHandler(context.Background(), nil, Options{})
	require.Error(t, err)
}

func Test_LimitRangeListHandler(t *testing.T) {
	now := testutil.Time()
	object := testutil.CreateLimitRange("limits")
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &corev1.LimitRangeList{Items: []corev1.LimitRange{*object}}

	cols := component.NewTableCols("Name", "Limits", "Age")

	runNativeListTest(t, nativeListCase{
		title:       "Limit Ranges",
		placeholder: "We couldn't find any limit ranges!",
		cols:        cols,
		object:      object,
		row: component.TableRow{
			"Limits": component.NewText("1"),
			"Age":    component.NewTimestamp(now),
		},
		status:   component.TextStatusOK,
		messages: []string{"v1 LimitRange is OK"},
		call: func(ctx context.Context, options Options) (component.Component, error) {
			return LimitRangeListHandler(ctx, list, options)
		},
	})

	_, err := LimitRangeListHandler(context.Background(), nil, Options{})
	require.Error(t, err)
}

func Test_LeaseListHandler(t *testing.T) {
	now := testutil.Time()
	object := testutil.CreateLease("lease")
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &coordinationv1.LeaseList{Items: []coordinationv1.Lease{*object}}

	cols := component.NewTableCols("Name", "Holder", "Age")

	runNativeListTest(t, nativeListCase{
		title:       "Leases",
		placeholder: "We couldn't find any leases!",
		cols:        cols,
		object:      object,
		row: component.TableRow{
			"Holder": component.NewText(""),
			"Age":    component.NewTimestamp(now),
		},
		status:   component.TextStatusOK,
		messages: []string{"coordination.k8s.io/v1 Lease is OK"},
		call: func(ctx context.Context, options Options) (component.Component, error) {
			return LeaseListHandler(ctx, list, options)
		},
	})

	_, err := LeaseListHandler(context.Background(), nil, Options{})
	require.Error(t, err)
}

func Test_CSIStorageCapacityListHandler(t *testing.T) {
	now := testutil.Time()
	object := testutil.CreateCSIStorageCapacity("capacity")
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &storagev1.CSIStorageCapacityList{Items: []storagev1.CSIStorageCapacity{*object}}

	cols := component.NewTableCols("Name", "Storage Class", "Capacity", "Age")

	runNativeListTest(t, nativeListCase{
		title:       "CSI Storage Capacities",
		placeholder: "We couldn't find any csi storage capacities!",
		cols:        cols,
		object:      object,
		row: component.TableRow{
			"Storage Class": component.NewText("manual"),
			"Capacity":      component.NewText("1Gi"),
			"Age":           component.NewTimestamp(now),
		},
		status:   component.TextStatusOK,
		messages: []string{"storage.k8s.io/v1 CSIStorageCapacity is OK"},
		call: func(ctx context.Context, options Options) (component.Component, error) {
			return CSIStorageCapacityListHandler(ctx, list, options)
		},
	})

	_, err := CSIStorageCapacityListHandler(context.Background(), nil, Options{})
	require.Error(t, err)
}

func Test_PriorityClassListHandler(t *testing.T) {
	now := testutil.Time()
	object := testutil.CreatePriorityClass("priority")
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &schedulingv1.PriorityClassList{Items: []schedulingv1.PriorityClass{*object}}

	cols := component.NewTableCols("Name", "Value", "Global Default", "Age")

	runNativeListTest(t, nativeListCase{
		title:       "Priority Classes",
		placeholder: "We couldn't find any priority classes!",
		cols:        cols,
		object:      object,
		row: component.TableRow{
			"Value":          component.NewText("1000"),
			"Global Default": component.NewText("false"),
			"Age":            component.NewTimestamp(now),
		},
		status:   component.TextStatusOK,
		messages: []string{"scheduling.k8s.io/v1 PriorityClass is OK"},
		call: func(ctx context.Context, options Options) (component.Component, error) {
			return PriorityClassListHandler(ctx, list, options)
		},
	})

	_, err := PriorityClassListHandler(context.Background(), nil, Options{})
	require.Error(t, err)
}

func Test_RuntimeClassListHandler(t *testing.T) {
	now := testutil.Time()
	object := testutil.CreateRuntimeClass("runtime")
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &nodev1.RuntimeClassList{Items: []nodev1.RuntimeClass{*object}}

	cols := component.NewTableCols("Name", "Handler", "Age")

	runNativeListTest(t, nativeListCase{
		title:       "Runtime Classes",
		placeholder: "We couldn't find any runtime classes!",
		cols:        cols,
		object:      object,
		row: component.TableRow{
			"Handler": component.NewText("runsc"),
			"Age":     component.NewTimestamp(now),
		},
		status:   component.TextStatusOK,
		messages: []string{"node.k8s.io/v1 RuntimeClass is OK"},
		call: func(ctx context.Context, options Options) (component.Component, error) {
			return RuntimeClassListHandler(ctx, list, options)
		},
	})

	_, err := RuntimeClassListHandler(context.Background(), nil, Options{})
	require.Error(t, err)
}

func Test_IngressClassListHandler(t *testing.T) {
	now := testutil.Time()
	object := testutil.CreateIngressClass("ingress")
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &networkingv1.IngressClassList{Items: []networkingv1.IngressClass{*object}}

	cols := component.NewTableCols("Name", "Controller", "Age")

	runNativeListTest(t, nativeListCase{
		title:       "Ingress Classes",
		placeholder: "We couldn't find any ingress classes!",
		cols:        cols,
		object:      object,
		row: component.TableRow{
			"Controller": component.NewText("k8s.io/ingress-nginx"),
			"Age":        component.NewTimestamp(now),
		},
		status:   component.TextStatusOK,
		messages: []string{"networking.k8s.io/v1 IngressClass is OK"},
		call: func(ctx context.Context, options Options) (component.Component, error) {
			return IngressClassListHandler(ctx, list, options)
		},
	})

	_, err := IngressClassListHandler(context.Background(), nil, Options{})
	require.Error(t, err)
}

func Test_CSIDriverListHandler(t *testing.T) {
	now := testutil.Time()
	object := testutil.CreateCSIDriver("driver")
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &storagev1.CSIDriverList{Items: []storagev1.CSIDriver{*object}}

	cols := component.NewTableCols("Name", "Attach Required", "Pod Info On Mount", "Age")

	runNativeListTest(t, nativeListCase{
		title:       "CSI Drivers",
		placeholder: "We couldn't find any csi drivers!",
		cols:        cols,
		object:      object,
		row: component.TableRow{
			"Attach Required":   component.NewText("true"),
			"Pod Info On Mount": component.NewText(""),
			"Age":               component.NewTimestamp(now),
		},
		status:   component.TextStatusOK,
		messages: []string{"storage.k8s.io/v1 CSIDriver is OK"},
		call: func(ctx context.Context, options Options) (component.Component, error) {
			return CSIDriverListHandler(ctx, list, options)
		},
	})

	_, err := CSIDriverListHandler(context.Background(), nil, Options{})
	require.Error(t, err)
}

func Test_CSINodeListHandler(t *testing.T) {
	now := testutil.Time()
	object := testutil.CreateCSINode("csi-node")
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &storagev1.CSINodeList{Items: []storagev1.CSINode{*object}}

	cols := component.NewTableCols("Name", "Drivers", "Age")

	runNativeListTest(t, nativeListCase{
		title:       "CSI Nodes",
		placeholder: "We couldn't find any csi nodes!",
		cols:        cols,
		object:      object,
		row: component.TableRow{
			"Drivers": component.NewText("driver"),
			"Age":     component.NewTimestamp(now),
		},
		status:   component.TextStatusOK,
		messages: []string{"storage.k8s.io/v1 CSINode is OK"},
		call: func(ctx context.Context, options Options) (component.Component, error) {
			return CSINodeListHandler(ctx, list, options)
		},
	})

	_, err := CSINodeListHandler(context.Background(), nil, Options{})
	require.Error(t, err)
}

func Test_VolumeAttachmentListHandler(t *testing.T) {
	now := testutil.Time()
	object := testutil.CreateVolumeAttachment("attachment")
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &storagev1.VolumeAttachmentList{Items: []storagev1.VolumeAttachment{*object}}

	cols := component.NewTableCols("Name", "Attacher", "Node", "Attached", "Age")

	runNativeListTest(t, nativeListCase{
		title:       "Volume Attachments",
		placeholder: "We couldn't find any volume attachments!",
		cols:        cols,
		object:      object,
		row: component.TableRow{
			"Attacher": component.NewText("attacher"),
			"Node":     component.NewText("node"),
			"Attached": component.NewText("true"),
			"Age":      component.NewTimestamp(now),
		},
		status:   component.TextStatusOK,
		messages: []string{"storage.k8s.io/v1 VolumeAttachment is OK"},
		call: func(ctx context.Context, options Options) (component.Component, error) {
			return VolumeAttachmentListHandler(ctx, list, options)
		},
	})

	_, err := VolumeAttachmentListHandler(context.Background(), nil, Options{})
	require.Error(t, err)
}

func Test_FlowSchemaListHandler(t *testing.T) {
	now := testutil.Time()
	object := testutil.CreateFlowSchema("flow")
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &flowcontrolv1.FlowSchemaList{Items: []flowcontrolv1.FlowSchema{*object}}

	cols := component.NewTableCols("Name", "Priority Level", "Matching Precedence", "Age")

	runNativeListTest(t, nativeListCase{
		title:       "Flow Schemas",
		placeholder: "We couldn't find any flow schemas!",
		cols:        cols,
		object:      object,
		row: component.TableRow{
			"Priority Level":      component.NewText("global-default"),
			"Matching Precedence": component.NewText("1000"),
			"Age":                 component.NewTimestamp(now),
		},
		status:   component.TextStatusOK,
		messages: []string{"flowcontrol.apiserver.k8s.io/v1 FlowSchema is OK"},
		call: func(ctx context.Context, options Options) (component.Component, error) {
			return FlowSchemaListHandler(ctx, list, options)
		},
	})

	_, err := FlowSchemaListHandler(context.Background(), nil, Options{})
	require.Error(t, err)
}

func Test_PriorityLevelConfigurationListHandler(t *testing.T) {
	now := testutil.Time()
	object := testutil.CreatePriorityLevelConfiguration("plc")
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &flowcontrolv1.PriorityLevelConfigurationList{Items: []flowcontrolv1.PriorityLevelConfiguration{*object}}

	cols := component.NewTableCols("Name", "Type", "Age")

	runNativeListTest(t, nativeListCase{
		title:       "Priority Level Configurations",
		placeholder: "We couldn't find any priority level configurations!",
		cols:        cols,
		object:      object,
		row: component.TableRow{
			"Type": component.NewText("Limited"),
			"Age":  component.NewTimestamp(now),
		},
		status:   component.TextStatusOK,
		messages: []string{"flowcontrol.apiserver.k8s.io/v1 PriorityLevelConfiguration is OK"},
		call: func(ctx context.Context, options Options) (component.Component, error) {
			return PriorityLevelConfigurationListHandler(ctx, list, options)
		},
	})

	_, err := PriorityLevelConfigurationListHandler(context.Background(), nil, Options{})
	require.Error(t, err)
}

func Test_ValidatingAdmissionPolicyListHandler(t *testing.T) {
	now := testutil.Time()
	object := testutil.CreateValidatingAdmissionPolicy("policy")
	object.CreationTimestamp = metav1.Time{Time: now}

	list := &admissionregistrationv1.ValidatingAdmissionPolicyList{
		Items: []admissionregistrationv1.ValidatingAdmissionPolicy{*object},
	}

	cols := component.NewTableCols("Name", "Failure Policy", "Age")

	runNativeListTest(t, nativeListCase{
		title:       "Validating Admission Policies",
		placeholder: "We couldn't find any validating admission policies!",
		cols:        cols,
		object:      object,
		row: component.TableRow{
			"Failure Policy": component.NewText(""),
			"Age":            component.NewTimestamp(now),
		},
		status:   component.TextStatusOK,
		messages: []string{"admissionregistration.k8s.io/v1 ValidatingAdmissionPolicy is OK"},
		call: func(ctx context.Context, options Options) (component.Component, error) {
			return ValidatingAdmissionPolicyListHandler(ctx, list, options)
		},
	})

	_, err := ValidatingAdmissionPolicyListHandler(context.Background(), nil, Options{})
	require.Error(t, err)
}
