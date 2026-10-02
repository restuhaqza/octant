package objectvisitor

import (
	"context"

	"github.com/pkg/errors"
	"go.opencensus.io/trace"
	"golang.org/x/sync/errgroup"
	autoscalingv1 "k8s.io/api/autoscaling/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"

	"github.com/vmware-tanzu/octant/internal/gvk"
	"github.com/vmware-tanzu/octant/internal/queryer"
	"github.com/vmware-tanzu/octant/internal/util/kubernetes"
)

// HorizontalPodAutoscaler is a typed visitor for horizontal pod autoscalers.
type HorizontalPodAutoscaler struct {
	queryer queryer.Queryer
	gvk     schema.GroupVersionKind
}

var _ TypedVisitor = (*HorizontalPodAutoscaler)(nil)

// NewHorizontalPodAutoscaler creates an instance of HorizontalPodAutoscaler for autoscaling/v1.
func NewHorizontalPodAutoscaler(q queryer.Queryer) *HorizontalPodAutoscaler {
	return &HorizontalPodAutoscaler{queryer: q, gvk: gvk.HorizontalPodAutoscaler}
}

// NewHorizontalPodAutoscalerV2 creates an instance of HorizontalPodAutoscaler for autoscaling/v2.
func NewHorizontalPodAutoscalerV2(q queryer.Queryer) *HorizontalPodAutoscaler {
	return &HorizontalPodAutoscaler{queryer: q, gvk: gvk.HorizontalPodAutoscalerV2}
}

// Supports returns the gvk this typed visitor supports.
func (h *HorizontalPodAutoscaler) Supports() schema.GroupVersionKind {
	return h.gvk
}

// Visit visits a hpa. It looks for an associated scale target (replication controllers, deployments, and replica sets)
func (s *HorizontalPodAutoscaler) Visit(ctx context.Context, object *unstructured.Unstructured, handler ObjectHandler, visitor Visitor, visitDescendants bool, level int) error {
	ctx, span := trace.StartSpan(ctx, "visitHorizontalPodAutoscaler")
	defer span.End()

	// The scale target reference has the same shape across autoscaling versions, so read it
	// from the unstructured object instead of converting to a version-specific type.
	scaleTargetRef, err := scaleTargetRefFromUnstructured(object)
	if err != nil {
		return err
	}

	hpa := &autoscalingv1.HorizontalPodAutoscaler{
		TypeMeta: metav1.TypeMeta{
			APIVersion: object.GetAPIVersion(),
			Kind:       object.GetKind(),
		},
		ObjectMeta: metav1.ObjectMeta{
			Namespace: object.GetNamespace(),
			Name:      object.GetName(),
		},
		Spec: autoscalingv1.HorizontalPodAutoscalerSpec{
			ScaleTargetRef: scaleTargetRef,
		},
	}
	level = handler.SetLevel(hpa.Kind, level)

	var g errgroup.Group

	g.Go(func() error {
		target, err := s.queryer.ScaleTarget(ctx, hpa)
		if err != nil {
			return err
		}

		if target != nil {
			g.Go(func() error {
				u := &unstructured.Unstructured{Object: target}
				if err := visitor.Visit(ctx, u, handler, true, level); err != nil {
					return errors.Wrapf(err, "horizontal pod scaler %s visit scale target",
						kubernetes.PrintObject(hpa))
				}

				return handler.AddEdge(ctx, object, u, level)
			})
		}

		return nil
	})

	return g.Wait()
}

// scaleTargetRefFromUnstructured reads spec.scaleTargetRef from an unstructured horizontal
// pod autoscaler. This works for both autoscaling/v1 and autoscaling/v2 objects.
func scaleTargetRefFromUnstructured(object *unstructured.Unstructured) (autoscalingv1.CrossVersionObjectReference, error) {
	if object == nil {
		return autoscalingv1.CrossVersionObjectReference{}, errors.New("horizontal pod autoscaler is nil")
	}

	kind, _, err := unstructured.NestedString(object.Object, "spec", "scaleTargetRef", "kind")
	if err != nil {
		return autoscalingv1.CrossVersionObjectReference{}, errors.Wrap(err, "get horizontal pod autoscaler scale target kind")
	}

	name, _, err := unstructured.NestedString(object.Object, "spec", "scaleTargetRef", "name")
	if err != nil {
		return autoscalingv1.CrossVersionObjectReference{}, errors.Wrap(err, "get horizontal pod autoscaler scale target name")
	}

	apiVersion, _, err := unstructured.NestedString(object.Object, "spec", "scaleTargetRef", "apiVersion")
	if err != nil {
		return autoscalingv1.CrossVersionObjectReference{}, errors.Wrap(err, "get horizontal pod autoscaler scale target api version")
	}

	return autoscalingv1.CrossVersionObjectReference{
		Kind:       kind,
		Name:       name,
		APIVersion: apiVersion,
	}, nil
}
