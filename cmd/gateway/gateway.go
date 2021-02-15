package main

import (
	"fmt"
	ht "html/template"
	"io"
	"mime"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strings"
	tt "text/template"
	"time"

	"anachronauts.club/repos/gmikit"
)

type Gateway struct {
	config       *GatewayConfig
	logger       *SplitLogger
	template     *ht.Template
	rootURL      *url.URL
	timeout      time.Duration
	imagePattern *regexp.Regexp
	externals    map[string]*tt.Template
}

func NewGateway(logger *SplitLogger, config *GatewayConfig) (*Gateway, error) {
	var err error
	g := &Gateway{
		config:    config,
		logger:    logger,
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

	g.template, err = loadTemplates(config.Templates)
	if err != nil {
		return nil, err
	}

	g.imagePattern, err = regexp.Compile(config.ImagePattern)
	if err != nil {
		return nil, err
	}

	return g, nil
}

func loadTemplates(templateDir string) (*ht.Template, error) {
	return ht.New("t").Funcs(ht.FuncMap{
		"safeURL": func(url *url.URL) ht.URL {
			return ht.URL(url.String())
		},
	}).ParseGlob(path.Join(templateDir, "*"))
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
			// TODO render better
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(w, "Gateway error: %v", err)
			return
		}
		defer resp.Close()

		switch resp.Status.Class() {
		// TODO: handle inputs
		//case gmikit.StatusClassInput:
		//	g.handleInput(w, r, resp, req)
		case gmikit.StatusClassSuccess:
			g.handleSuccess(w, r, resp, req)
		case gmikit.StatusClassRedirect:
			g.handleRedirect(w, r, resp, req)
		case gmikit.StatusClassTemporaryFailure:
			g.handleTemporaryFailure(w, r, resp, req)
		case gmikit.StatusClassPermanentFailure:
			g.handlePermanentFailure(w, r, resp, req)
		default:
			g.handleUnknown(w, r, resp, req)
		}

	case "POST":
		r.ParseForm()
		if q, ok := r.Form["q"]; !ok {
			// TODO template
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Bad request"))
		} else {
			// TODO template
			w.Header().Add("Location", "?"+q[0])
			w.WriteHeader(http.StatusFound)
			w.Write([]byte("Redirecting..."))
		}

	default:
		// TODO template
		status := http.StatusMethodNotAllowed
		w.WriteHeader(status)
		fmt.Fprintf(w, "%d %s", status, http.StatusText(status))
	}
}

func (g *Gateway) render(
	w http.ResponseWriter,
	template string,
	httpStatus int,
	ctx interface{},
) {
	w.Header().Add("Content-Type", "text/html")
	w.WriteHeader(httpStatus)
	if err := g.template.ExecuteTemplate(w, template, ctx); err != nil {
		g.logger.Error("Failed to execute template:", err)
		// At this point it's kinda too late to recover anything, since we've
		// probably spewed a bunch of stuff out over the connection.
	}
}

func (g *Gateway) showError(
	w http.ResponseWriter,
	resp *gmikit.Response,
	req *gmikit.Request,
	httpStatus int,
	message string,
) {
	ctx := &ErrorContext{
		RenderContext: RenderContext{
			Request:  req,
			Response: resp,
		},
		HTTPStatus: http.StatusBadGateway,
		Message:    message,
	}
	g.render(w, "error.html", ctx.HTTPStatus, ctx)
}

func (g *Gateway) handleInput(
	w http.ResponseWriter,
	r *http.Request,
	resp *gmikit.Response,
	req *gmikit.Request,
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
		g.showError(
			w, resp, req,
			http.StatusBadGateway,
			fmt.Sprintf("Error parsing mime-type \"%s\": %v", resp.Meta, err),
		)
		return
	}

	if m != "text/gemini" {
		w.Header().Add("Content-Type", resp.Meta)
		io.Copy(w, resp.Body)
		return
	}

	// Build render context
	rc := NewSuccessContext(
		g.imagePattern,
		func(url *url.URL) (*url.URL, string, error) {
			target, err := g.convertURL(url, r.URL)
			class := url.Scheme
			if !url.IsAbs() || url.Host == g.rootURL.Host {
				if url.Scheme == "" {
					class = "local gemini"
				} else {
					class = fmt.Sprintf("local %s", url.Scheme)
				}
			}
			return target, class, err
		})
	gmikit.ParseLines(resp.Body, rc)
	rc.Request = req
	rc.Response = resp

	g.render(w, "2x.html", http.StatusOK, rc)
}

func (g *Gateway) handleRedirect(
	w http.ResponseWriter,
	r *http.Request,
	resp *gmikit.Response,
	req *gmikit.Request,
) {
	to, err := url.Parse(resp.Meta)
	if err != nil {
		g.showError(
			w, resp, req,
			http.StatusBadGateway,
			fmt.Sprintf("While parsing redirect url: %v", err),
		)
		return
	}
	next := req.URL.ResolveReference(to)
	autoRedirect := next.Scheme == "gemini"
	next, err = g.convertURL(next, r.URL)
	if err != nil {
		g.logger.Errorf(
			"Error converting URL (%s, %s): %v", next, r.URL, err)
		g.showError(
			w, resp, req,
			http.StatusInternalServerError,
			fmt.Sprintf("Error converting URL \"%s\" for redirect: %v",
				next,
				err))
		return
	}

	w.Header().Add("Content-Type", "text/html")
	if autoRedirect {
		http.Redirect(w, r, next.String(), http.StatusFound)
	}
	ctx := &RedirectContext{
		RenderContext: RenderContext{
			Request:  req,
			Response: resp,
		},
		Location: next,
	}
	if err := g.template.ExecuteTemplate(w, "3x.html", ctx); err != nil {
		g.logger.Error("Failed to execute template:", err)
		// At this point it's kinda too late to recover anything, since we've
		// probably spewed a bunch of stuff out over the connection.
	}
}

func (g *Gateway) handleTemporaryFailure(
	w http.ResponseWriter,
	r *http.Request,
	resp *gmikit.Response,
	req *gmikit.Request,
) {
	ctx := &ErrorContext{
		RenderContext: RenderContext{
			Request:  req,
			Response: resp,
		},
	}

	switch resp.Status {
	default:
		ctx.HTTPStatus = http.StatusBadGateway
	case gmikit.StatusServerUnavailable:
		ctx.HTTPStatus = http.StatusServiceUnavailable
	case gmikit.StatusCGIError:
		ctx.HTTPStatus = http.StatusBadGateway
	case gmikit.StatusProxyError:
		ctx.HTTPStatus = http.StatusBadGateway
	case gmikit.StatusSlowDown:
		w.Header().Add("Retry-After", resp.Meta)
		ctx.HTTPStatus = http.StatusTooManyRequests
	}

	g.render(w, "4x.html", ctx.HTTPStatus, ctx)
}

func (g *Gateway) handlePermanentFailure(
	w http.ResponseWriter,
	r *http.Request,
	resp *gmikit.Response,
	req *gmikit.Request,
) {
	ctx := &ErrorContext{
		RenderContext: RenderContext{
			Request:  req,
			Response: resp,
		},
	}

	switch resp.Status {
	default:
		ctx.HTTPStatus = http.StatusForbidden
	case gmikit.StatusNotFound:
		ctx.HTTPStatus = http.StatusNotFound
	case gmikit.StatusGone:
		ctx.HTTPStatus = http.StatusGone
	case gmikit.StatusProxyRequestRefused:
		ctx.HTTPStatus = http.StatusForbidden
	case gmikit.StatusBadRequest:
		ctx.HTTPStatus = http.StatusBadGateway
		ctx.Message = "Bad request (this might be the fault of the server)"
		g.logger.Errorf("Sent bad request, meta=%s", resp.Meta)
	}

	g.render(w, "5x.html", ctx.HTTPStatus, ctx)
}

func (g *Gateway) handleUnknown(
	w http.ResponseWriter,
	r *http.Request,
	resp *gmikit.Response,
	req *gmikit.Request,
) {
	ctx := &ErrorContext{
		RenderContext: RenderContext{
			Request:  req,
			Response: resp,
		},
		HTTPStatus: http.StatusNotImplemented,
	}
	g.render(w, "unknown.html", ctx.HTTPStatus, ctx)
}
