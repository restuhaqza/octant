package objectvisitor_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/vmware-tanzu/octant/internal/gvk"
	"github.com/vmware-tanzu/octant/internal/objectvisitor"
	"github.com/vmware-tanzu/octant/internal/objectvisitor/fake"
	queryerFake "github.com/vmware-tanzu/octant/internal/queryer/fake"
	"github.com/vmware-tanzu/octant/internal/testutil"
)

func TestHorizontalPodAutoscalerV2_Supports(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	q := queryerFake.NewMockQueryer(controller)

	assert.Equal(t, gvk.HorizontalPodAutoscaler, objectvisitor.NewHorizontalPodAutoscaler(q).Supports())
	assert.Equal(t, gvk.HorizontalPodAutoscalerV2, objectvisitor.NewHorizontalPodAutoscalerV2(q).Supports())
}

func TestHorizontalPodAutoscalerV2_Visit(t *testing.T) {
	controller := gomock.NewController(t)
	defer controller.Finish()

	object := &autoscalingv2.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "autoscaling/v2",
			Kind:       "HorizontalPodAutoscaler",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "hpa",
			Namespace: "default",
			UID:       "hpa-uid",
		},
		Spec: autoscalingv2.HorizontalPodAutoscalerSpec{
			MaxReplicas: 10,
			ScaleTargetRef: autoscalingv2.CrossVersionObjectReference{
				APIVersion: "apps/v1",
				Kind:       "Deployment",
				Name:       "deployment",
			},
		},
	}
	u := testutil.ToUnstructured(t, object)

	deployment := testutil.CreateDeployment("deployment")
	deploymentUnstructured := testutil.ToUnstructured(t, deployment)

	q := queryerFake.NewMockQueryer(controller)
	q.EXPECT().
		ScaleTarget(gomock.Any(), gomock.Any()).
		Return(deploymentUnstructured.Object, nil)

	handler := fake.NewMockObjectHandler(controller)
	handler.EXPECT().SetLevel("HorizontalPodAutoscaler", 1).Return(2)
	handler.EXPECT().
		AddEdge(gomock.Any(), u, &unstructured.Unstructured{Object: deploymentUnstructured.Object}, 2).
		Return(nil)

	visitor := fake.NewMockVisitor(controller)
	visitor.EXPECT().
		Visit(gomock.Any(), gomock.Any(), handler, true, 2).
		Return(nil)

	ctx := context.Background()

	hpaVisitor := objectvisitor.NewHorizontalPodAutoscalerV2(q)
	require.NoError(t, hpaVisitor.Visit(ctx, u, handler, visitor, true, 1))
}
