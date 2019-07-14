package content

import (
	"fmt"
	"net/http"
	"strings"
)

var (
	prefix = []string{
		"/assets",
		"/translations",
		"/fonts",
	}
	paths = []string{
		"/humans.txt",
		"/index.html",
		"/",
		"/robots.txt",
	}
)

func Handler() http.Handler {
	static := staticHandler()

	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		if isIndexRequest(req.URL.Path) {
			req.URL.Path = "/"
		}
		fmt.Println(req.Method, req.URL.Path)
		static.ServeHTTP(rw, req)
	})
}

func isIndexRequest(path string) bool {
	for _, prefix := range prefix {
		if strings.HasPrefix(path, prefix) {
			return false
		}
	}

	for _, p := range paths {
		if path == p {
			return false
		}
	}

	return true
}
