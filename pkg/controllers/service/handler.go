package service

import (
	"context"

	"github.com/rancher/k3v/pkg/translate"
	v1 "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/apply"
	corev1 "k8s.io/api/core/v1"
)

const (
	vServiceNames = "vServiceNames"
)

type handler struct {
	targetNamespace string
	apply           apply.Apply

	vServiceCache v1.ServiceCache
	vService      v1.ServiceClient
	pService      v1.ServiceClient
}

func Register(
	ctx context.Context,
	targetNamespace string,
	apply apply.Apply,
	pService v1.ServiceController,
	vService v1.ServiceController,
) {

	h := &handler{
		targetNamespace: targetNamespace,
		apply:           apply,
		vServiceCache:   vService.Cache(),
		vService:        vService,
		pService:        pService,
	}

	vService.Cache().AddIndexer(vServiceNames, func(obj *corev1.Service) (strings []string, e error) {
		return []string{translate.ObjectToPName(obj)}, nil
	})

	vService.OnRemove(ctx, "service-populate", h.Remove)
	vService.OnChange(ctx, "service-populate", h.Populate)
	pService.OnChange(ctx, "service-backpopulate", h.BackPopulate)
}
