package main

import (
	"html/template"
	"strings"

	"anachronauts.club/repos/gmikit"
)

type RenderContext struct {
	gmikit.HtmlWriter
	Request     *gmikit.Request
	Response    *gmikit.Response
	Title       string
	titleLevel  int
	bodyBuilder *strings.Builder
}

func NewRenderContext(rewriter gmikit.UrlRewriter) *RenderContext {
	var builder strings.Builder
	return &RenderContext{
		HtmlWriter:  *gmikit.NewHtmlWriter(&builder, rewriter),
		titleLevel:  4, // gemini only supports 3 levels
		bodyBuilder: &builder,
	}
}

func (ctx *RenderContext) Body() template.HTML {
	return template.HTML(ctx.bodyBuilder.String())
}

func (ctx *RenderContext) Heading1(text string) error {
	if ctx.titleLevel > 1 {
		ctx.Title = text
		ctx.titleLevel = 1
	}
	return ctx.HtmlWriter.Heading1(text)
}

func (ctx *RenderContext) Heading2(text string) error {
	if ctx.titleLevel > 2 {
		ctx.Title = text
		ctx.titleLevel = 2
	}
	return ctx.HtmlWriter.Heading2(text)
}

func (ctx *RenderContext) Heading3(text string) error {
	if ctx.titleLevel > 3 {
		ctx.Title = text
		ctx.titleLevel = 3
	}
	return ctx.HtmlWriter.Heading3(text)
}
