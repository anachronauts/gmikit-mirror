package gmikit

import (
	"strings"
	"testing"
)

func TestConvertEmpty(t *testing.T) {
	expected := ""
	input := strings.NewReader(expected)
	var output strings.Builder
	if err := GmiToHtml(input, &output); err != nil {
		t.Error(err)
	}

	actual := output.String()
	if expected != actual {
		t.Errorf("Expected %v got %v", expected, actual)
	}
}

func TestConvertComplex(t *testing.T) {
	input := strings.NewReader(`# Heading 1

Text, more text.

=> hello.gmi Hello!

> One
> Two
> Three

` +
		"```a\n" +
		"tested\n" +
		"```b\n")

	expected := `<h1>Heading 1</h1>
<p>

Text, more text.

</p>
<a href="hello.gmi">Hello!</a><br/>
<p>

</p>
<blockquote>
One
Two
Three
</blockquote>
<p>

</p>
<div aria-label="a">
<pre aria-hidden="true" alt="a">
tested
</pre>
</div>
`

	var output strings.Builder
	if err := GmiToHtml(input, &output); err != nil {
		t.Error(err)
	}

	actual := output.String()
	if expected != actual {
		t.Errorf("Expected %v got %v", expected, actual)
	}
}
