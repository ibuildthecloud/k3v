package endpoint

import (
	"github.com/rancher/k3v/pkg/translate"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *handler) BackPopulate(key string, obj *v1.Endpoints) (*v1.Endpoints, error) {
	if obj == nil {
		objs, _ := h.vEndpointsCache.GetByIndex(vEndpointsNames, key)
		for _, obj := range objs {
			return obj, h.vEndpoints.Delete(obj.Namespace, obj.Name, nil)
		}
		return obj, nil
	} else if !translate.IsManaged(obj) {
		return obj, nil
	}

	_, err := h.vEndpointsCache.Get(translate.GetOwner(obj))
	if errors.IsNotFound(err) {
		return nil, h.pEndpoints.Delete(obj.Namespace, obj.Name, &metav1.DeleteOptions{})
	}

	return obj, nil
}
