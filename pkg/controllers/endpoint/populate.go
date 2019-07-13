package endpoint

import (
	"github.com/rancher/k3v/pkg/translate"
	corev1 "k8s.io/api/core/v1"
)

func (h *handler) Remove(key string, endpoints *corev1.Endpoints) (*corev1.Endpoints, error) {
	return endpoints, h.apply.WithOwner(endpoints).ApplyObjects()
}

func (h *handler) Populate(key string, endpoints *corev1.Endpoints) (*corev1.Endpoints, error) {
	if endpoints == nil {
		return nil, nil
	}

	newObj, err := translate.SetupMetadata(h.targetNamespace, endpoints)
	if err != nil {
		return nil, err
	}
	return endpoints, h.apply.WithOwner(endpoints).ApplyObjects(newObj)
}
