package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strconv"
	"strings"
	"syscall"
	"time"

	flag "github.com/spf13/pflag"
	"github.com/urfave/negroni"
)

var confDir string = "/usr/local/etc/gmikit"
var dataDir string = "/usr/local/share/gmikit"
var configFile *string = flag.StringP(
	"config", "c",
	path.Join(confDir, "gateway.conf"),
	"Path to config",
)
var templateDir = path.Join(dataDir, "templates")

func isProcessRunning(pid int) bool {
	// Always works on UNIX
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// boop the process and see if it exists
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		return false
	}

	return true
}

func WritePidFile(path string) error {
	// Check for an existing PID file
	pidStr, err := ioutil.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	if err == nil {
		// We have an existing PID file. Check if it's valid.
		pid, err := strconv.Atoi(strings.TrimSpace(string(pidStr)))
		if err == nil && isProcessRunning(pid) {
			return fmt.Errorf("Already running as pid %d", pid)
		}
	}

	// Either we don't have a PID file, or we've established it's stale.
	return ioutil.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0o644)
}

func main() {
	flag.Parse()

	// Parse config
	config, err := LoadConfig(*configFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Create logger
	logger, err := config.MakeLoggers()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating loggers: %v\n", err)
		os.Exit(1)
	}

	// Handle SIGHUP by reopening log files
	sighup := make(chan os.Signal, 1)
	signal.Notify(sighup, syscall.SIGHUP)
	go (func() {
		for {
			<-sighup
			fmt.Fprintln(os.Stderr, "Got SIGHUP, reopening logs")
			logger.Reopen()
		}
	})()

	// Start handling signals, then plonk down a PID file
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	if config.PidFile != "" {
		err := WritePidFile(config.PidFile)
		if err != nil {
			logger.Fatal(err)
		}
		defer os.RemoveAll(config.PidFile)
	}

	// Create handlers
	gateway, err := NewGateway(logger, config)
	if err != nil {
		logger.Fatal(err)
	}
	mux := http.NewServeMux()
	mux.Handle("/", gateway)
	mux.Handle("/favicon.ico", http.NotFoundHandler())

	// Create middleware
	n := negroni.New()
	n.Use(logger)
	n.UseHandler(mux)

	// Create and start server
	srv := &http.Server{
		Addr:    config.Bind,
		Handler: n,
	}
	go (func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal(err)
		}
	})()
	logger.Request("Started HTTP server on", config.Bind)

	<-done
	logger.Request("Stopping HTTP server")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer (func() {
		// extra handling here
		cancel()
	})()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server shutdown failed:", err)
	}
	logger.Request("Stopped HTTP server ok")
}
