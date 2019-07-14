package server

import (
	"net/http"

	"github.com/rancher/magellan/pkg/content"
	"github.com/rancher/magellan/pkg/proxy"
)

func Handler(kubeConfig string) (http.Handler, error) {
	mux := http.NewServeMux()
	p, err := proxy.HandlerFromConfig("/k8s-api", kubeConfig)
	if err != nil {
		return nil, err
	}

	mux.Handle("/k8s-api/", p)
	mux.Handle("/", content.Handler())

	return mux, nil
}
