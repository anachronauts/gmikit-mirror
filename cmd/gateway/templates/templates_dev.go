// +build dev

package templates

import (
	"net/http"
	"os"

	"github.com/shurcooL/httpfs/filter"
)

var Assets http.FileSystem = filter.Skip(
	http.Dir("assets"),
	func(path string, fi os.FileInfo) bool {
		return !fi.IsDir() && fi.Name() == "COPYING"
	})
