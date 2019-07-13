package proxy

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/rancher/k3v/pkg/translate"
	"github.com/rancher/magellan/pkg/proxy"
	"k8s.io/apiserver/pkg/server"
	"k8s.io/client-go/rest"
)

func NewBuildHandlerChain(namespace string, cfg *rest.Config) (func(apiHandler http.Handler, c *server.Config) http.Handler, error) {
	proxy, err := Handler(namespace, cfg)
	if err != nil {
		return nil, err
	}

	return func(apiHandler http.Handler, c *server.Config) http.Handler {
		mux := mux.NewRouter()
		mux.NotFoundHandler = apiHandler

		mux.Path("/api/v1/namespaces/{namespace}/{resource:pods}/{name}/{action:log}").Handler(proxy)
		mux.Path("/api/v1/namespaces/{namespace}/{resource:pods}/{name}/{action:attach}").Handler(proxy)
		mux.Path("/api/v1/namespaces/{namespace}/{resource:pods}/{name}/{action:exec}").Handler(proxy)
		mux.Path("/api/v1/namespaces/{namespace}/{resource:pods}/{name}/{action:portforward}").Handler(proxy)
		mux.Path("/api/v1/namespaces/{namespace}/{resource:pods}/{name}/{action:proxy}").Handler(proxy)
		mux.Path("/api/v1/namespaces/{namespace}/{resource:services}/{name}/{action:proxy}").Handler(proxy)
		mux.Path("/api/v1/{resource:nodes}/{name}/{action:proxy}").Handler(proxy)

		return server.DefaultBuildHandlerChain(mux, c)
	}, nil
}

func Handler(namespace string, cfg *rest.Config) (http.Handler, error) {
	next, err := proxy.Handler("", cfg)
	if err != nil {
		return nil, err
	}

	return &handler{
		targetNamespace: namespace,
		next:            next,
	}, nil
}

type handler struct {
	targetNamespace string
	next            http.Handler
}

func (h *handler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	resource := vars["resource"]
	name := vars["name"]
	namespace := vars["namespace"]
	action := vars["action"]

	targetName := translate.ToPName(name, namespace)
	if namespace == "" {
		req.URL.Path = fmt.Sprintf("/api/v1/%s/%s/%s", resource, targetName, action)
	} else {
		req.URL.Path = fmt.Sprintf("/api/v1/namespaces/%s/%s/%s/%s", h.targetNamespace, resource, targetName, action)
	}
	req.Header.Del("Authorization")
	h.next.ServeHTTP(rw, req)
}
