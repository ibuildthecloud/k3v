package configmap

import (
	"github.com/rancher/k3v/pkg/translate"
	corev1 "k8s.io/api/core/v1"
)

func (h *handler) Remove(key string, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	return configMap, h.apply.WithOwner(configMap).ApplyObjects()
}

func (h *handler) Populate(key string, configMap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	if configMap == nil {
		return nil, nil
	}

	newObj, err := translate.SetupMetadata(h.targetNamespace, configMap)
	if err != nil {
		return nil, err
	}
	return configMap, h.apply.WithOwner(configMap).ApplyObjects(newObj)
}
