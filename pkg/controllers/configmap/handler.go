package configmap

import (
	"context"

	"github.com/rancher/k3v/pkg/translate"
	v1 "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/apply"
	corev1 "k8s.io/api/core/v1"
)

const (
	vConfigMapNames = "vConfigMapNames"
)

type handler struct {
	targetNamespace string
	apply           apply.Apply

	vConfigMapCache v1.ConfigMapCache
	vConfigMap      v1.ConfigMapClient
	pConfigMap      v1.ConfigMapClient
}

func Register(
	ctx context.Context,
	targetNamespace string,
	apply apply.Apply,
	pConfigMap v1.ConfigMapController,
	vConfigMap v1.ConfigMapController,
) {
	h := &handler{
		targetNamespace: targetNamespace,
		apply:           apply,
		vConfigMapCache: vConfigMap.Cache(),
		vConfigMap:      vConfigMap,
		pConfigMap:      pConfigMap,
	}

	vConfigMap.Cache().AddIndexer(vConfigMapNames, func(obj *corev1.ConfigMap) (strings []string, e error) {
		return []string{translate.ObjectToPName(obj)}, nil
	})

	vConfigMap.OnRemove(ctx, "configMap-populate", h.Remove)
	vConfigMap.OnChange(ctx, "configMap-populate", h.Populate)
	pConfigMap.OnChange(ctx, "configMap-backpopulate", h.BackPopulate)
}
