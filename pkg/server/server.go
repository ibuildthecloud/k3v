package server

import (
	"os"

	"github.com/rancher/k3s/pkg/cli/cmds"
	"github.com/rancher/k3s/pkg/cli/server"
	"github.com/rancher/k3v/pkg/cluster"
	"github.com/rancher/k3v/pkg/controllers"
	"github.com/rancher/k3v/pkg/proxy"
	"github.com/rancher/k3v/pkg/translate"
	"github.com/rancher/wrangler/pkg/start"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	apiserver "k8s.io/apiserver/pkg/server"
)

type Config struct {
	Debug      bool
	Version    string
	KubeConfig string
	Namespace  string
	ListenPort int
	ID         string
}

func startK8S(ctx context.Context, cfg Config) (string, error) {
	serverConfig := &cmds.Server{
		DataDir:          "./k3v-data",
		KubeConfigOutput: "../../kubeconfig.yaml",
		TLSSan:           []string{"10.43.194.102"},
		ClusterCIDR:      "10.44.0.0/16",
		ServiceCIDR:      "10.45.0.0/16",
		ClusterDomain:    "cluster.local",
		HTTPSPort:        cfg.ListenPort,
		DisableAgent:     true,
		DisableScheduler: true,
		NoDeploy: []string{
			"traefik",
			"servicelb",
		},
		ExtraControllerArgs: []string{
			"controllers=*,-nodeipam,-nodelifecycle,-persistentvolume-binder,-attachdetach,-persistentvolume-expander",
		},
	}

	go func() {
		err := server.Start(ctx, cfg.Debug, cfg.Version, serverConfig, cmds.AgentConfig)
		logrus.Fatalf("k3s server stopped: %v", err)
	}()

	first := true
	for {
		if _, err := os.Stat(serverConfig.KubeConfigOutput); err == nil {
			break
		} else {
			if !first {
				logrus.Infof("waiting for %s to exist", serverConfig.KubeConfigOutput)
				first = true
			}
		}
	}

	return serverConfig.KubeConfigOutput, nil
}

func Run(ctx context.Context, cfg Config) error {
	translate.Generation = cfg.ID

	pContext, err := cluster.NewContext(cfg.KubeConfig, cfg.Namespace)
	if err != nil {
		return err
	}

	buildHandler, err := proxy.NewBuildHandlerChain(cfg.Namespace, pContext.RestConfig)
	if err != nil {
		return err
	}

	apiserver.OverrideBuildChainFunc = buildHandler

	vKubeConfig, err := startK8S(ctx, cfg)
	if err != nil {
		return err
	}

	vContext, err := cluster.NewContext(vKubeConfig, "")
	if err != nil {
		return err
	}

	controllers.Register(ctx, cfg.ListenPort, cfg.Namespace, pContext, vContext)

	if err := start.All(ctx, 5, append(pContext.Starters, vContext.Starters...)...); err != nil {
		return err
	}

	<-ctx.Done()
	return nil
}
