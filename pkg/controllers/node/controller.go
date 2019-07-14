package node

import (
	"context"
	"fmt"

	corev1controllers "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Register(
	ctx context.Context,
	vPod corev1controllers.PodController,
	pNodes corev1controllers.NodeController,
	vNodes corev1controllers.NodeController,
) {
	h := handler{
		podCache:  vPod.Cache(),
		nodeCache: vNodes.Cache(),
		nodes:     vNodes,
	}

	pNodes.OnChange(ctx, "node-backpopulate", h.OnChange)
	vPod.Cache().AddIndexer("assigned", func(pod *corev1.Pod) (strings []string, e error) {
		if pod.Spec.NodeName == "" {
			return nil, nil
		}
		return []string{pod.Spec.NodeName}, nil
	})
}

type handler struct {
	podCache  corev1controllers.PodCache
	nodeCache corev1controllers.NodeCache
	nodes     corev1controllers.NodeClient
}

func (h *handler) OnChange(key string, node *corev1.Node) (*corev1.Node, error) {
	if node == nil {
		return nil, nil
	}

	pods, err := h.podCache.GetByIndex("assigned", node.Name)
	if err != nil {
		return nil, err
	}
	if len(pods) == 0 {
		return node, nil
	}

	vNode, err := h.nodeCache.Get(node.Name)
	if errors.IsNotFound(err) {
		_, err := h.nodes.Create(&corev1.Node{
			ObjectMeta: metav1.ObjectMeta{
				Name: node.Name,
			},
		})
		if err != nil {
			return node, err
		}
		return node, fmt.Errorf("resync")
	} else if err != nil {
		return node, err
	}

	if !equality.Semantic.DeepEqual(vNode.Status, node.Status) {
		newNode := vNode.DeepCopy()
		newNode.Status = node.Status
		vNode, err = h.nodes.UpdateStatus(newNode)
		if err != nil {
			return node, err
		}
	}

	if !equality.Semantic.DeepEqual(vNode.Spec, node.Spec) {
		newNode := vNode.DeepCopy()
		newNode.Spec = node.Spec
		_, err = h.nodes.Update(newNode)
	}

	return node, err
}
