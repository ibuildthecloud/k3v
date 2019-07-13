package pod

import (
	"github.com/rancher/k3v/pkg/translate"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *handler) BackPopulate(key string, obj *v1.Pod) (*v1.Pod, error) {
	if obj == nil {
		objs, _ := h.vPodCache.GetByIndex(vPodNames, key)
		for _, obj := range objs {
			return obj, h.vPods.Delete(obj.Namespace, obj.Name, &metav1.DeleteOptions{
				GracePeriodSeconds: &zero,
			})
		}
		return obj, nil
	} else if !translate.IsManaged(obj) {
		return obj, nil
	}

	vPod, err := h.vPodCache.Get(translate.GetOwner(obj))
	if errors.IsNotFound(err) {
		return nil, h.pPods.Delete(obj.Namespace, obj.Name, &metav1.DeleteOptions{})
	}
	if err != nil {
		return vPod, err
	}

	if obj.DeletionTimestamp != nil {
		if vPod.DeletionTimestamp == nil {
			return obj, h.vPods.Delete(vPod.Namespace, vPod.Name, &metav1.DeleteOptions{
				GracePeriodSeconds: &zero,
			})
		}
		return obj, nil
	}

	if vPod.Spec.NodeName != obj.Spec.NodeName {
		err := h.k8sVPods.Pods(vPod.Namespace).Bind(&v1.Binding{
			ObjectMeta: metav1.ObjectMeta{
				Name:      vPod.Name,
				Namespace: vPod.Namespace,
			},
			Target: v1.ObjectReference{
				Kind:       "Node",
				Name:       obj.Spec.NodeName,
				APIVersion: "v1",
			},
		})
		if err != nil {
			return vPod, err
		}
	}

	if !equality.Semantic.DeepEqual(vPod.Status, obj.Status) {
		newPod := vPod.DeepCopy()
		newPod.Status = obj.Status
		return h.vPods.UpdateStatus(newPod)
	}

	return obj, nil
}
