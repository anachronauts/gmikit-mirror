package main

import (
	"log"
	"net/http"
	"path"
	"time"

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
		start := time.Now()
		next.ServeHTTP(w, r)
		elapsed := time.Since(start)
		log.Println(r.Method, r.URL.Path, elapsed)
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
