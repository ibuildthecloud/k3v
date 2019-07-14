// +build !release

package content

import "net/http"

func staticHandler() http.Handler {
	return http.FileServer(http.Dir("./dist"))
}
