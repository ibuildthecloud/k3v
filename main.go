package main

import (
	"context"
	"fmt"
	"os"

	"github.com/rancher/k3v/pkg/server"
	"github.com/rancher/wrangler/pkg/signals"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

var (
	Version   = "v0.0.0-dev"
	GitCommit = "HEAD"
	config    server.Config
)

func main() {
	app := cli.NewApp()
	app.Name = "k3v"
	app.Version = fmt.Sprintf("%s (%s)", Version, GitCommit)
	app.Usage = "Virtual Kubernetes"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "kubeconfig",
			EnvVar:      "KUBECONFIG",
			Destination: &config.KubeConfig,
		},
		cli.StringFlag{
			Name:        "id",
			EnvVar:      "K3V_ID",
			Value:       "1",
			Destination: &config.ID,
		},
		cli.StringFlag{
			Name:        "namespace",
			EnvVar:      "NAMESPACE",
			Value:       "default",
			Destination: &config.Namespace,
		},
		cli.IntFlag{
			Name:        "listen-port",
			EnvVar:      "K3V_LISTEN_PORT",
			Value:       7443,
			Destination: &config.ListenPort,
		},
		cli.BoolFlag{
			Name: "debug",
		},
	}
	app.Action = run

	if err := app.Run(os.Args); err != nil {
		logrus.Fatal(err)
	}
}

func run(c *cli.Context) error {
	if c.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}

	logrus.Info("Starting controller")
	ctx := signals.SetupSignalHandler(context.Background())

	return server.Run(ctx, config)
}
