//go:generate go run pkg/codegen/main.go

package main

import (
	"context"
	"flag"

	"github.com/rancher/wrangler/pkg/signals"
	"github.com/rancher/wrangler/pkg/start"
	"github.com/rancher/wrangler/sample/pkg/generated/controllers/apps"
	"github.com/rancher/wrangler/sample/pkg/generated/controllers/samplecontroller.k8s.io"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	masterURL  string
	kubeconfig string
)

func main() {
	flag.Parse()

	// set up signals so we handle the first shutdown signal gracefully
	ctx := signals.SetupSignalHandler(context.Background())

	cfg, err := clientcmd.BuildConfigFromFlags(masterURL, kubeconfig)
	if err != nil {
		logrus.Fatalf("Error building kubeconfig: %s", err.Error())
	}

	kubeClient, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		logrus.Fatalf("Error building kubernetes clientset: %s", err.Error())
	}

	apps, err := apps.NewFactoryFromConfig(cfg)
	if err != nil {
		logrus.Fatalf("Error building apps controllers: %s", err.Error())
	}

	sample, err := samplecontroller.NewFactoryFromConfig(cfg)
	if err != nil {
		logrus.Fatalf("Error building sample controllers: %s", err.Error())
	}

	Register(ctx, kubeClient, apps.Apps().V1().Deployment(), sample.Samplecontroller().V1alpha1().Foo())

	if err := start.All(ctx, 2, apps, sample); err != nil {
		logrus.Fatalf("Error starting: %s", err.Error())
	}

	<-ctx.Done()
}

func init() {
	flag.StringVar(&kubeconfig, "kubeconfig", "", "Path to a kubeconfig. Only required if out-of-cluster.")
	flag.StringVar(&masterURL, "master", "", "The address of the Kubernetes API server. Overrides any value in kubeconfig. Only required if out-of-cluster.")
}
