package secret

import (
	"github.com/rancher/k3v/pkg/translate"
	corev1 "k8s.io/api/core/v1"
)

func (h *handler) Remove(key string, secrets *corev1.Secret) (*corev1.Secret, error) {
	return secrets, h.apply.WithOwner(secrets).ApplyObjects()
}

func (h *handler) Populate(key string, secrets *corev1.Secret) (*corev1.Secret, error) {
	if secrets == nil {
		return nil, nil
	}

	newObj, err := translate.SetupMetadata(h.targetNamespace, secrets)
	if err != nil {
		return nil, err
	}
	newSecret := newObj.(*corev1.Secret)
	newSecret.Type = corev1.SecretTypeOpaque
	return secrets, h.apply.WithOwner(secrets).ApplyObjects(newSecret)
}
