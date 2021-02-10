package main

import (
	"crypto/x509"
	"encoding/base64"
	"io"
	"log"
	"net/url"
	"os"

	"anachronauts.club/repos/gmikit"
	flag "github.com/spf13/pflag"
	"golang.org/x/crypto/blake2b"
)

var output *string = flag.StringP("output", "o", "-", "Output path")
var redirect *int = flag.IntP("redirect", "r", 5, "Maximum number of redirects")

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

	if flag.NArg() != 1 {
		log.Fatalf("usage: %s [options] url", os.Args[0])
	}
	url, err := url.Parse(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}

	client := &gmikit.Client{
		TrustCertificate: func(hostname string, cert *x509.Certificate) error {
			fingerprint := blake2b.Sum512(cert.Raw)
			log.Println("Fingerprint", hostname, base64.StdEncoding.EncodeToString(fingerprint[:]))
			return nil
		},
	}

	req := gmikit.NewRequest(url)
	for {
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Close()

		switch resp.Status.Class() {
		case gmikit.StatusClassSuccess:
			_, err := io.Copy(w, resp.Body)
			if err != nil {
				log.Fatal(err)
			}
			os.Exit(0)

		case gmikit.StatusClassRedirect:
			log.Println(resp.Status, resp.Meta)
			url, err := url.Parse(resp.Meta)
			if err != nil {
				log.Fatal(err)
			}
			req.URL = url
			if *redirect > 0 {
				*redirect--
			} else {
				os.Exit(int(resp.Status))
			}

		default:
			log.Print(resp.Status)
			os.Exit(int(resp.Status))
		}
	}
}
