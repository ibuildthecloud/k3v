package cluster

import (
	"time"

	"k8s.io/client-go/rest"

	"github.com/pkg/errors"
	"github.com/rancher/wrangler-api/pkg/generated/controllers/core"
	corev1 "github.com/rancher/wrangler-api/pkg/generated/controllers/core/v1"
	"github.com/rancher/wrangler/pkg/apply"
	"github.com/rancher/wrangler/pkg/kubeconfig"
	"github.com/rancher/wrangler/pkg/start"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
)

type Context struct {
	RestConfig *rest.Config
	K8s        kubernetes.Interface
	Core       corev1.Interface
	Apply      apply.Apply
	Starters   []start.Starter
}

func NewContext(kubeConfig, namespace string) (*Context, error) {
	for {
		cfg, err := kubeconfig.GetNonInteractiveClientConfig(kubeConfig).ClientConfig()
		if err != nil {
			return nil, errors.Wrap(err, "Error building kubeconfig")
		}

		k8s, err := kubernetes.NewForConfig(cfg)
		if err != nil {
			return nil, errors.Wrap(err, "Error building client")
		}

		first := true
		_, err = k8s.Discovery().ServerVersion()
		if err != nil {
			if !first {
				logrus.Infof("Waiting on kubernetes at %s: %v", cfg.Host, err)
			}
			first = false
			time.Sleep(2 * time.Second)
			continue
		}

		apply := apply.New(k8s.Discovery(), apply.NewClientFactory(cfg))

		controllers, err := core.NewFactoryFromConfigWithNamespace(cfg, namespace)
		if err != nil {
			return nil, errors.Wrap(err, "Error building controllers")
		}

		return &Context{
			RestConfig: cfg,
			K8s:        k8s,
			Core:       controllers.Core().V1(),
			Apply: apply.WithCacheTypes(
				controllers.Core().V1().Service(),
				controllers.Core().V1().Pod(),
				controllers.Core().V1().Service(),
				controllers.Core().V1().Endpoints(),
				controllers.Core().V1().ConfigMap(),
				controllers.Core().V1().Secret(),
			).WithStrictCaching(),
			Starters: []start.Starter{
				controllers,
			},
		}, nil
	}
}
