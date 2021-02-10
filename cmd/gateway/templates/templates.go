package templates

//go:generate go run -tags=dev templates_generate.go

import (
	"html/template"
	"net/http"
	"net/url"

	"github.com/shurcooL/httpfs/html/vfstemplate"
)

type union struct {
	fs []http.FileSystem
}

func (u *union) Open(name string) (http.File, error) {
	var err error
	for _, fs := range u.fs {
		var file http.File
		file, err = fs.Open(name)
		if err == nil {
			return file, nil
		}
	}
	return nil, err
}

func LoadFS(path string) http.FileSystem {
	if path != "" {
		u := &union{fs: make([]http.FileSystem, 2)}
		u.fs[0] = Assets
		u.fs[1] = http.Dir(path)
		return u
	} else {
		return Assets
	}
}

func Load(path string) (*template.Template, error) {
	t := template.New("t").Funcs(template.FuncMap{
		"safeURL": func(url *url.URL) template.URL {
			return template.URL(url.String())
		},
	})
	return vfstemplate.ParseGlob(LoadFS(path), t, "*")
}
