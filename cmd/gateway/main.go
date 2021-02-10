package main

import (
	"log"
	"net/http"
	"path"

	flag "github.com/spf13/pflag"
)

var confDir string = "/usr/local/etc/gmikit"
var configFile *string = flag.StringP(
	"config", "c",
	path.Join(confDir, "gateway.conf"),
	"Path to config",
)

func requestLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func main() {
	flag.Parse()

	config, err := LoadConfig(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	gateway, err := NewGateway(config)
	if err != nil {
		log.Fatal(err)
	}

	http.Handle("/", requestLogger(gateway))
	http.Handle("/favicon.ico", requestLogger(http.NotFoundHandler()))
	log.Printf("HTTP server listening on %s", config.Bind)
	log.Fatal(http.ListenAndServe(config.Bind, nil))
}
