package configmap

import (
	"github.com/rancher/k3v/pkg/translate"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *handler) BackPopulate(key string, obj *v1.ConfigMap) (*v1.ConfigMap, error) {
	if obj == nil {
		objs, _ := h.vConfigMapCache.GetByIndex(vConfigMapNames, key)
		for _, obj := range objs {
			return obj, h.vConfigMap.Delete(obj.Namespace, obj.Name, nil)
		}
		return obj, nil
	} else if !translate.IsManaged(obj) {
		return obj, nil
	}

	_, err := h.vConfigMapCache.Get(translate.GetOwner(obj))
	if errors.IsNotFound(err) {
		return nil, h.pConfigMap.Delete(obj.Namespace, obj.Name, &metav1.DeleteOptions{})
	}

	return obj, nil
}
