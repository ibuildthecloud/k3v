// +build release

package content

import (
	"net/http"

	assetfs "github.com/elazarl/go-bindata-assetfs"
	"github.com/rancher/magellan/pkg/bindata"
)

func staticHandler() http.Handler {
	return http.FileServer(&assetfs.AssetFS{
		Prefix:    "dist",
		Asset:     bindata.Asset,
		AssetDir:  bindata.AssetDir,
		AssetInfo: bindata.AssetInfo,
	})
}
