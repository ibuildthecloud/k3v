package endpoint

import (
	"context"

	"github.com/rancher/k3v/pkg/translate"
	v1 "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/apply"
	corev1 "k8s.io/api/core/v1"
)

const (
	vEndpointsNames = "vEndpointsNames"
)

type handler struct {
	targetNamespace string
	apply           apply.Apply

	vEndpointsCache v1.EndpointsCache
	vEndpoints      v1.EndpointsClient
	pEndpoints      v1.EndpointsClient
}

func Register(
	ctx context.Context,
	targetNamespace string,
	apply apply.Apply,
	pEndpoints v1.EndpointsController,
	vEndpoints v1.EndpointsController,
) {

	h := &handler{
		targetNamespace: targetNamespace,
		apply:           apply,
		pEndpoints:      pEndpoints,
		vEndpointsCache: vEndpoints.Cache(),
		vEndpoints:      vEndpoints,
	}

	vEndpoints.Cache().AddIndexer(vEndpointsNames, func(obj *corev1.Endpoints) (strings []string, e error) {
		return []string{translate.ObjectToPName(obj)}, nil
	})

	vEndpoints.OnRemove(ctx, "endpoint-populate", h.Remove)
	vEndpoints.OnChange(ctx, "endpoint-populate", h.Populate)
	pEndpoints.OnChange(ctx, "endpoint-backpopulate", h.BackPopulate)
}
