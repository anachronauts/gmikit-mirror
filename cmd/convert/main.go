package main

import (
	"io"
	"log"
	"os"

	"anachronauts.club/repos/gmikit"
	flag "github.com/spf13/pflag"
)

var format *string = flag.StringP("format", "T", "html", "Output format (gmi, html)")
var output *string = flag.StringP("output", "o", "-", "Output path")

func main() {
	flag.Parse()

	var w io.Writer = os.Stdout
	if *output != "-" {
		var err error
		w, err = os.Create(*output)
		if err != nil {
			log.Fatal(err)
		}
	}

	var v gmikit.Visitor
	switch *format {
	case "gmi":
		v = gmikit.NewGmiWriter(w)
	case "html":
		v = gmikit.NewHtmlWriter(w, nil)
	default:
		log.Fatalf("unknown format '%v'", *format)
	}

	if flag.NArg() == 0 {
		err := gmikit.ParseLines(os.Stdin, v)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		for _, arg := range flag.Args() {
			r, err := os.Open(arg)
			if err != nil {
				log.Fatal(err)
			}

			err = gmikit.ParseLines(r, v)
			if err != nil {
				log.Fatal(err)
			}
		}
	}
}
