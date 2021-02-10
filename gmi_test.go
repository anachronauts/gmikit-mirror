package gmikit

import (
	"strings"
	"testing"
)

func TestRoundtripEmpty(t *testing.T) {
	expected := ""
	input := strings.NewReader(expected)
	var output strings.Builder
	if err := NormalizeGmi(input, &output); err != nil {
		t.Error(err)
	}

	actual := output.String()
	if expected != actual {
		t.Errorf("Expected %v got %v", expected, actual)
	}
}

func TestRoundtripComplex(t *testing.T) {
	expected := `# Heading 1

Text, more text.

=> hello.gmi Hello!

> One
> Two
> Three
`

	input := strings.NewReader(expected)
	var output strings.Builder
	if err := NormalizeGmi(input, &output); err != nil {
		t.Error(err)
	}

	actual := output.String()
	if expected != actual {
		t.Errorf("Expected %v got %v", expected, actual)
	}
}
