package service

import (
	"github.com/rancher/k3v/pkg/translate"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *handler) BackPopulate(key string, obj *v1.Service) (*v1.Service, error) {
	if obj == nil {
		objs, _ := h.vServiceCache.GetByIndex(vServiceNames, key)
		for _, obj := range objs {
			return obj, h.vService.Delete(obj.Namespace, obj.Name, nil)
		}
		return obj, nil
	} else if !translate.IsManaged(obj) {
		return obj, nil
	}

	vSvc, err := h.vServiceCache.Get(translate.GetOwner(obj))
	if errors.IsNotFound(err) {
		return nil, h.pService.Delete(obj.Namespace, obj.Name, &metav1.DeleteOptions{})
	}
	if err != nil {
		return obj, err
	}

	if vSvc.Spec.ClusterIP != obj.Spec.ClusterIP || !equality.Semantic.DeepEqual(vSvc.Spec.Ports, obj.Spec.Ports) {
		newService := vSvc.DeepCopy()
		newService.Spec.ClusterIP = obj.Spec.ClusterIP
		newService.Spec.Ports = obj.Spec.Ports
		vSvc, err = h.vService.Update(newService)
		if err != nil {
			return obj, err
		}
	}

	if !equality.Semantic.DeepEqual(vSvc.Status, obj.Status) {
		newService := vSvc.DeepCopy()
		newService.Status = obj.Status
		vSvc, err = h.vService.UpdateStatus(newService)
		if err != nil {
			return obj, err
		}
	}

	return obj, nil
}
