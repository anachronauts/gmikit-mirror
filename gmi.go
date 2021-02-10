package gmikit

import (
	"bufio"
	"fmt"
	"io"
	"net/url"
	"strings"
)

type Visitor interface {
	Begin() error
	End() error
	Text(text string) error
	Link(target *url.URL, friendlyName string) error
	PreformattingToggle(altText string) error
	PreformattedText(text string) error
	Heading1(text string) error
	Heading2(text string) error
	Heading3(text string) error
	UnorderedListItem(text string) error
	Quote(text string) error
}

func ParseLines(r io.Reader, v Visitor) error {
	const ws = " \t"
	pre := false
	scanner := bufio.NewScanner(r)
	if err := v.Begin(); err != nil {
		return err
	}
	for scanner.Scan() {
		text := scanner.Text()
		if strings.HasPrefix(text, "```") {
			pre = !pre
			if err := v.PreformattingToggle(text[3:]); err != nil {
				return err
			}
		} else if pre {
			if err := v.PreformattedText(text); err != nil {
				return err
			}
		} else if strings.HasPrefix(text, "=>") {
			text = text[2:]
			text = strings.TrimLeft(text, ws)
			split := strings.IndexAny(text, ws)
			if split == -1 {
				url, err := url.Parse(text)
				if err != nil {
					return err
				}

				if err = v.Link(url, ""); err != nil {
					return err
				}
			} else {
				url, err := url.Parse(text[:split])
				if err != nil {
					return err
				}

				name := strings.TrimLeft(text[split:], ws)
				if err = v.Link(url, name); err != nil {
					return err
				}
			}
		} else if strings.HasPrefix(text, "*") {
			text = strings.TrimLeft(text[1:], ws)
			if err := v.UnorderedListItem(text); err != nil {
				return err
			}
		} else if strings.HasPrefix(text, "###") {
			text = strings.TrimLeft(text[3:], ws)
			if err := v.Heading3(text); err != nil {
				return err
			}
		} else if strings.HasPrefix(text, "##") {
			text = strings.TrimLeft(text[2:], ws)
			if err := v.Heading2(text); err != nil {
				return err
			}
		} else if strings.HasPrefix(text, "#") {
			text = strings.TrimLeft(text[1:], ws)
			if err := v.Heading1(text); err != nil {
				return err
			}
		} else if strings.HasPrefix(text, ">") {
			text = strings.TrimLeft(text[1:], ws)
			if err := v.Quote(text); err != nil {
				return err
			}
		} else {
			if err := v.Text(text); err != nil {
				return err
			}
		}
	}
	if err := v.End(); err != nil {
		return err
	}

	return scanner.Err()
}

type GmiWriter struct {
	w   io.Writer
	pre bool
}

func NewGmiWriter(w io.Writer) *GmiWriter {
	return &GmiWriter{w: w, pre: false}
}

func (g *GmiWriter) Begin() error { return nil }
func (g *GmiWriter) End() error   { return nil }

func (g *GmiWriter) Text(text string) error {
	_, err := fmt.Fprintf(g.w, "%s\n", text)
	return err
}

func (g *GmiWriter) Link(target *url.URL, friendlyName string) error {
	if friendlyName == "" {
		_, err := fmt.Fprintf(g.w, "=> %s\n", target)
		return err
	} else {
		_, err := fmt.Fprintf(g.w, "=> %s %s\n", target, friendlyName)
		return err
	}
}

func (g *GmiWriter) PreformattingToggle(altText string) error {
	if g.pre {
		g.pre = false
		_, err := fmt.Fprintf(g.w, "```\n")
		return err
	} else {
		g.pre = true
		_, err := fmt.Fprintf(g.w, "```%s\n", altText)
		return err
	}
}

func (g *GmiWriter) PreformattedText(text string) error {
	_, err := fmt.Fprintf(g.w, "%s\n", text)
	return err
}

func (g *GmiWriter) Heading1(text string) error {
	_, err := fmt.Fprintf(g.w, "# %s\n", text)
	return err
}

func (g *GmiWriter) Heading2(text string) error {
	_, err := fmt.Fprintf(g.w, "## %s\n", text)
	return err
}

func (g *GmiWriter) Heading3(text string) error {
	_, err := fmt.Fprintf(g.w, "### %s\n", text)
	return err
}

func (g *GmiWriter) UnorderedListItem(text string) error {
	_, err := fmt.Fprintf(g.w, "* %s\n", text)
	return err
}

func (g *GmiWriter) Quote(text string) error {
	_, err := fmt.Fprintf(g.w, "> %s\n", text)
	return err
}

func NormalizeGmi(r io.Reader, w io.Writer) error {
	return ParseLines(r, NewGmiWriter(w))
}
