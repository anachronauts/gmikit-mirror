package main

import (
	"html/template"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"anachronauts.club/repos/gmikit"
	"anachronauts.club/repos/gmikit/cmd/gateway/theme"
)

var userPatterns = []*regexp.Regexp{
	regexp.MustCompile(`~([^/?]+)`),
	regexp.MustCompile(`(?i)/users/([^/?]+)`),
}

type RenderContext struct {
	Request  *gmikit.Request
	Response *gmikit.Response
	style    *Style
}

type SuccessContext struct {
	RenderContext
	gmikit.HtmlWriter
	Title       string
	titleLevel  int
	bodyBuilder *strings.Builder
}

type RedirectContext struct {
	RenderContext
	Location *url.URL
}

type ErrorContext struct {
	RenderContext
	HTTPStatus int
	Message    string
}

type Style struct {
	Light *theme.Theme
	Dark  *theme.Theme
}

func (ctx *RenderContext) Site() string {
	// We define "site" in a tilde-friendly way, similar to lagrange. So first,
	// we look for user names.
	for _, re := range userPatterns {
		if m := re.FindStringSubmatch(ctx.Request.URL.Path); m != nil {
			return m[1]
		}
	}

	// If there are no usernames, then we use the hostname
	return ctx.Request.URL.Hostname()
}

func (ctx *RenderContext) Style() *Style {
	if ctx.style == nil {
		site := ctx.Site()
		ctx.style = &Style{
			Light: theme.NewWhiteTheme(site),
			Dark:  theme.NewColorfulDarkTheme(site),
		}
	}
	return ctx.style
}

func NewSuccessContext(rewriter gmikit.UrlRewriter) *SuccessContext {
	ctx := &SuccessContext{
		titleLevel:  4, // gemini only supports 3 levels
		bodyBuilder: &strings.Builder{},
	}
	ctx.HtmlWriter = *gmikit.NewHtmlWriter(ctx.bodyBuilder, rewriter)
	return ctx
}

func (ctx *SuccessContext) Body() template.HTML {
	return template.HTML(ctx.bodyBuilder.String())
}

func (ctx *SuccessContext) Heading1(text string) error {
	if ctx.titleLevel > 1 {
		ctx.Title = text
		ctx.titleLevel = 1
	}
	return ctx.HtmlWriter.Heading1(text)
}

func (ctx *SuccessContext) Heading2(text string) error {
	if ctx.titleLevel > 2 {
		ctx.Title = text
		ctx.titleLevel = 2
	}
	return ctx.HtmlWriter.Heading2(text)
}

func (ctx *SuccessContext) Heading3(text string) error {
	if ctx.titleLevel > 3 {
		ctx.Title = text
		ctx.titleLevel = 3
	}
	return ctx.HtmlWriter.Heading3(text)
}

func (ctx *ErrorContext) HTTPFriendly() string {
	return http.StatusText(ctx.HTTPStatus)
}
