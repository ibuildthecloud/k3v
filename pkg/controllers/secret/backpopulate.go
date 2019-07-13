package secret

import (
	"github.com/rancher/k3v/pkg/translate"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func (h *handler) BackPopulate(key string, obj *v1.Secret) (*v1.Secret, error) {
	if obj == nil {
		objs, _ := h.vSecretCache.GetByIndex(vSecretNames, key)
		for _, obj := range objs {
			return obj, h.vSecret.Delete(obj.Namespace, obj.Name, nil)
		}
		return obj, nil
	} else if !translate.IsManaged(obj) {
		return obj, nil
	}

	_, err := h.vSecretCache.Get(translate.GetOwner(obj))
	if errors.IsNotFound(err) {
		return nil, h.pSecret.Delete(obj.Namespace, obj.Name, &metav1.DeleteOptions{})
	}

	return obj, nil
}
