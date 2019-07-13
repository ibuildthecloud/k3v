package secret

import (
	"context"

	"github.com/rancher/k3v/pkg/translate"
	v1 "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/apply"
	corev1 "k8s.io/api/core/v1"
)

const (
	vSecretNames = "vSecretNames"
)

type handler struct {
	targetNamespace string
	apply           apply.Apply

	vSecretCache v1.SecretCache
	vSecret      v1.SecretClient
	pSecret      v1.SecretClient
}

func Register(
	ctx context.Context,
	targetNamespace string,
	apply apply.Apply,
	pSecret v1.SecretController,
	vSecret v1.SecretController,
) {

	h := &handler{
		targetNamespace: targetNamespace,
		apply:           apply,
		pSecret:         pSecret,
		vSecretCache:    vSecret.Cache(),
		vSecret:         vSecret,
	}

	vSecret.Cache().AddIndexer(vSecretNames, func(obj *corev1.Secret) (strings []string, e error) {
		return []string{translate.ObjectToPName(obj)}, nil
	})

	vSecret.OnRemove(ctx, "secret-populate", h.Remove)
	vSecret.OnChange(ctx, "secret-populate", h.Populate)
	pSecret.OnChange(ctx, "secret-backpopulate", h.BackPopulate)
}
