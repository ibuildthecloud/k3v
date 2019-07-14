package controllers

import (
	"context"

	"github.com/rancher/k3v/pkg/cluster"
	"github.com/rancher/k3v/pkg/controllers/configmap"
	"github.com/rancher/k3v/pkg/controllers/endpoint"
	"github.com/rancher/k3v/pkg/controllers/node"
	"github.com/rancher/k3v/pkg/controllers/pod"
	"github.com/rancher/k3v/pkg/controllers/secret"
	"github.com/rancher/k3v/pkg/controllers/service"
)

func Register(ctx context.Context, port int, targetNamespace string, pContext, vContext *cluster.Context) {
	configmap.Register(ctx, targetNamespace, pContext.Apply, pContext.Core.ConfigMap(), vContext.Core.ConfigMap())
	endpoint.Register(ctx, targetNamespace, pContext.Apply, pContext.Core.Endpoints(), vContext.Core.Endpoints())
	node.Register(ctx, vContext.Core.Pod(), pContext.Core.Node(), vContext.Core.Node())
	pod.Register(ctx, targetNamespace, pContext.Apply, pContext.Core.Pod(), vContext.Core.Pod(), vContext.K8s.CoreV1(), pContext.Core.Service(), port)
	secret.Register(ctx, targetNamespace, pContext.Apply, pContext.Core.Secret(), vContext.Core.Secret())
	service.Register(ctx, targetNamespace, pContext.Apply, pContext.Core.Service(), vContext.Core.Service())
}
