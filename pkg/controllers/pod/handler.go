package pod

import (
	"context"
	"fmt"
	"sync"

	"github.com/rancher/k3v/pkg/translate"
	v1 "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/apply"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	typedcorev1 "k8s.io/client-go/kubernetes/typed/core/v1"
)

const (
	ClusterDomain = "cluster.local"
	vPodNames     = "vPodNames"
)

var (
	F = false
)

type handler struct {
	sync.Mutex

	targetNamespace string
	apply           apply.Apply

	vPodCache     v1.PodCache
	vPods         v1.PodClient
	pPods         v1.PodClient
	pPodCache     v1.PodCache
	pServiceCache v1.ServiceCache
	k8sVPods      typedcorev1.PodsGetter
	port          int

	pinged bool
}

func Register(
	ctx context.Context,
	targetNamespace string,
	apply apply.Apply,
	pPods v1.PodController,
	vPods v1.PodController,
	k8sVPods typedcorev1.PodsGetter,
	pServices v1.ServiceController,
	port int,
) {

	h := &handler{
		targetNamespace: targetNamespace,
		apply: apply.WithPatcher(corev1.SchemeGroupVersion.WithKind("Pod"), func(namespace, name string, pt types.PatchType, data []byte) (runtime.Object, error) {
			err := pPods.Delete(namespace, name, &metav1.DeleteOptions{})
			if err == nil {
				return nil, fmt.Errorf("replace pod")
			}
			return nil, err
		}),
		pPods:         pPods,
		pPodCache:     pPods.Cache(),
		vPods:         vPods,
		vPodCache:     vPods.Cache(),
		k8sVPods:      k8sVPods,
		pServiceCache: pServices.Cache(),
		port:          port,
	}

	vPods.Cache().AddIndexer(vPodNames, func(obj *corev1.Pod) (strings []string, e error) {
		return []string{translate.ObjectToPName(obj)}, nil
	})

	vPods.OnRemove(ctx, "pod-populate", h.Remove)
	vPods.OnChange(ctx, "pod-populate", h.Populate)
	pPods.OnChange(ctx, "pod-backpopulate", h.BackPopulate)
}
