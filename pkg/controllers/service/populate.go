package service

import (
	"github.com/rancher/k3v/pkg/translate"
	corev1 "k8s.io/api/core/v1"
)

func (h *handler) Remove(key string, services *corev1.Service) (*corev1.Service, error) {
	return services, h.apply.WithOwner(services).ApplyObjects()
}

func (h *handler) Populate(key string, services *corev1.Service) (*corev1.Service, error) {
	if services == nil {
		return nil, nil
	}

	newObj, err := translate.SetupMetadata(h.targetNamespace, services)
	if err != nil {
		return nil, err
	}

	newService := newObj.(*corev1.Service)
	newService.Spec.Selector = nil
	newService.Spec.ClusterIP = ""
	for i := range newService.Spec.Ports {
		newService.Spec.Ports[i].NodePort = 0
	}

	return services, h.apply.WithOwner(services).ApplyObjects(newService)
}
