package gmikit

import (
	"fmt"
	"html/template"
	"io"
	"net/url"
)

type UrlRewriter func(*url.URL) (*url.URL, string, error)

type element int

const (
	Clear element = iota
	Pre
	AltPre
	Quoting
	Para
	List
)

type HtmlWriter struct {
	w        io.Writer
	Rewriter UrlRewriter
	elem     element
}

func NewHtmlWriter(w io.Writer, rewriter UrlRewriter) *HtmlWriter {
	return &HtmlWriter{
		w:        w,
		Rewriter: rewriter,
		elem:     Clear,
	}
}

func (h *HtmlWriter) Write(p []byte) (n int, err error) {
	return h.w.Write(p)
}

func (h *HtmlWriter) Clear() error {
	var err error
	switch h.elem {
	case Pre:
		_, err = fmt.Fprintf(h.w, "</pre>\n")
	case AltPre:
		_, err = fmt.Fprintf(h.w, "</pre>\n</div>\n")
	case Quoting:
		_, err = fmt.Fprintf(h.w, "</blockquote>\n")
	case Para:
		_, err = fmt.Fprintf(h.w, "</p>\n")
	case List:
		_, err = fmt.Fprint(h.w, "</ul>\n")
	}
	if err != nil {
		return err
	}
	h.elem = Clear
	return nil
}

func (h *HtmlWriter) Begin() error { return nil }

func (h *HtmlWriter) End() error {
	return h.Clear()
}

func (h *HtmlWriter) Text(text string) error {
	if h.elem != Para {
		err := h.Clear()
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintf(h.w, "<p>"); err != nil {
			return err
		}

		h.elem = Para
	}

	_, err := fmt.Fprintf(h.w, "%s\n", template.HTMLEscapeString(text))
	return err
}

var anchor = template.Must(
	template.New("anchor").Parse(
		"<a href=\"{{.Href}}\"" +
			"{{ if .Class }} class=\"{{.Class}}\" {{- end }}>" +
			"{{.Text}}</a><br/>\n"))

func (h *HtmlWriter) Link(target *url.URL, friendlyName string) error {
	err := h.Clear()
	if err != nil {
		return err
	}

	class := target.Scheme
	if h.Rewriter != nil {
		target, class, err = h.Rewriter(target)
		if err != nil {
			return err
		}
	}

	if friendlyName == "" {
		friendlyName = target.String()
	}

	return anchor.Execute(h.w, struct {
		Href  *url.URL
		Class string
		Text  string
	}{
		Href:  target,
		Class: class,
		Text:  friendlyName,
	})
}

var altPre = template.Must(
	template.New("altPre").
		Parse(`<div aria-label="{{.}}">
<pre aria-hidden="true" alt="{{.}}">
`))

func (h *HtmlWriter) PreformattingToggle(altText string) error {
	elem := h.elem
	err := h.Clear()
	if err != nil {
		return err
	}

	if elem == Pre || elem == AltPre {
		return nil
	}

	if altText == "" {
		h.elem = Pre
		_, err := fmt.Fprint(h.w, "<pre>\n")
		return err
	} else {
		h.elem = AltPre
		return altPre.Execute(h.w, altText)
	}
}

func (h *HtmlWriter) PreformattedText(text string) error {
	_, err := fmt.Fprintf(h.w, "%s\n", template.HTMLEscapeString(text))
	return err
}

var h1 = template.Must(
	template.New("h1").
		Parse("<h1>{{.}}</h1>\n"))

func (h *HtmlWriter) Heading1(text string) error {
	err := h.Clear()
	if err != nil {
		return err
	}

	return h1.Execute(h.w, text)
}

var h2 = template.Must(
	template.New("h2").
		Parse("<h2>{{.}}</h2>\n"))

func (h *HtmlWriter) Heading2(text string) error {
	err := h.Clear()
	if err != nil {
		return err
	}

	return h2.Execute(h.w, text)
}

var h3 = template.Must(
	template.New("h3").
		Parse("<h3>{{.}}</h3>\n"))

func (h *HtmlWriter) Heading3(text string) error {
	err := h.Clear()
	if err != nil {
		return err
	}

	return h3.Execute(h.w, text)
}

var li = template.Must(
	template.New("li").
		Parse("<li>{{.}}</li>\n"))

func (h *HtmlWriter) UnorderedListItem(text string) error {
	if h.elem != List {
		err := h.Clear()
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintf(h.w, "<ul>\n"); err != nil {
			return err
		}
		h.elem = List
	}

	return li.Execute(h.w, text)
}

func (h *HtmlWriter) Quote(text string) error {
	if h.elem != Quoting {
		err := h.Clear()
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintf(h.w, "<blockquote>"); err != nil {
			return err
		}
		h.elem = Quoting
	}

	_, err := fmt.Fprintf(h.w, "%s\n", template.HTMLEscapeString(text))
	return err
}

func GmiToHtml(r io.Reader, w io.Writer) error {
	return ParseLines(r, NewHtmlWriter(w, nil))
}
