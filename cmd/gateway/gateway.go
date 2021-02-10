package main

import (
	"fmt"
	ht "html/template"
	"io"
	"log"
	"mime"
	"net/http"
	"net/url"
	"strings"
	tt "text/template"
	"time"

	"anachronauts.club/repos/gmikit"
	"anachronauts.club/repos/gmikit/cmd/gateway/templates"
)

type Gateway struct {
	config    *GatewayConfig
	template  *ht.Template
	rootURL   *url.URL
	timeout   time.Duration
	externals map[string]*tt.Template
}

func NewGateway(config *GatewayConfig) (*Gateway, error) {
	var err error
	g := &Gateway{
		config:    config,
		timeout:   time.Duration(config.Timeout) * time.Millisecond,
		externals: make(map[string]*tt.Template),
	}

	g.rootURL, err = url.Parse(config.Root)
	if err != nil {
		return nil, err
	}

	for k, v := range config.External {
		t, err := tt.New(k).Parse(v)
		if err != nil {
			return nil, err
		}
		g.externals[k] = t
	}

	g.template, err = templates.Load(config.Templates)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func (g *Gateway) convertURL(
	target *url.URL,
	requestBase *url.URL,
) (*url.URL, error) {
	if !target.IsAbs() {
		return target, nil
	}

	if target.Scheme == "gemini" && target.Host == g.rootURL.Host {
		// This is an internal URL
		out := *target
		out.Scheme = requestBase.Scheme
		out.Host = requestBase.Host
		return &out, nil
	}

	// External URL
	tmpl, ok := g.externals[target.Scheme]
	if !ok {
		tmpl, ok = g.externals["_"]
		if !ok {
			return target, nil
		}
	}
	var out strings.Builder
	if err := tmpl.Execute(&out, target); err != nil {
		return nil, err
	}
	return url.Parse(out.String())
}

func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		client := gmikit.Client{Timeout: g.timeout}
		req := gmikit.NewRequest(&url.URL{
			Scheme:   g.rootURL.Scheme,
			Host:     g.rootURL.Host,
			Path:     r.URL.Path,
			RawQuery: r.URL.RawQuery,
		})
		resp, err := client.Do(req)
		if err != nil {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(w, "Gateway error: %v", err)
			return
		}
		defer resp.Close()

		switch resp.Status.Class() {
		case gmikit.StatusClassInput:
			g.handleInput(w, resp, r)
		case gmikit.StatusClassSuccess:
			g.handleSuccess(w, r, resp, req)
		case gmikit.StatusClassRedirect:
			g.handleRedirect(w, r, resp, req)
		case gmikit.StatusClassTemporaryFailure:
		case gmikit.StatusClassPermanentFailure:
		default:
		}

	case "POST":
		r.ParseForm()
		if q, ok := r.Form["q"]; !ok {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad request"))
		} else {
			w.Header().Add("Location", "?"+q[0])
			w.WriteHeader(http.StatusFound)
			w.Write([]byte("Redirecting..."))
		}

	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Write([]byte("404 not found"))
	}
}

func (g *Gateway) handleInput(
	w http.ResponseWriter,
	resp *gmikit.Response,
	req *http.Request,
) {
	// TODO
}

func (g *Gateway) handleSuccess(
	w http.ResponseWriter,
	r *http.Request,
	resp *gmikit.Response,
	req *gmikit.Request,
) {
	m, _, err := mime.ParseMediaType(resp.Meta)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(
			w, "Gateway error: %s %s: %v",
			resp.Status,
			resp.Meta,
			err,
		)
		return
	}

	if m != "text/gemini" {
		w.Header().Add("Content-Type", resp.Meta)
		io.Copy(w, resp.Body)
		return
	}

	// Build render context
	rc := NewRenderContext(func(url *url.URL) (*url.URL, string, error) {
		target, err := g.convertURL(url, r.URL)
		class := url.Scheme
		if !url.IsAbs() || url.Host == g.rootURL.Host {
			class = "local"
		}
		return target, class, err
	})
	gmikit.ParseLines(resp.Body, rc)
	rc.Request = req
	rc.Response = resp

	w.Header().Add("Content-Type", "text/html")
	if err := g.template.ExecuteTemplate(w, "2x.html", rc); err != nil {
		log.Print(err)
	}
}

func (g *Gateway) handleRedirect(
	w http.ResponseWriter,
	r *http.Request,
	resp *gmikit.Response,
	req *gmikit.Request,
) {
	to, err := url.Parse(resp.Meta)
	if err != nil {
		// TODO template
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "Gateway error: bad redirect %v", err)
		return
	}
	next := req.URL.ResolveReference(to)
	autoRedirect := next.Scheme == "gemini"
	next, err = g.convertURL(next, r.URL)
	if err != nil {
		// TODO template
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprintf(w, "Gateway error: bad template %v", err)
		return
	}

	if !autoRedirect {
		// TODO template
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "This page is redirecting you to %s", next)
		return
	}

	w.Header().Add("Location", next.String())
	w.WriteHeader(http.StatusFound)
	fmt.Fprintf(w, "Redirecting to %s", next)
}
