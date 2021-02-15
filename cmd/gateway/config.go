package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/client9/reopen"
	"github.com/pelletier/go-toml"
	"github.com/urfave/negroni"
)

type GatewayConfig struct {
	Bind         string            `toml:"bind"`
	Root         string            `toml:"root"`
	Timeout      int64             `toml:"timeout"`
	Templates    string            `toml:"templates"`
	RequestLog   string            `toml:"request_log"`
	ErrorLog     string            `toml:"error_log"`
	PidFile      string            `toml:"pid_file"`
	ImagePattern string            `toml:"image_pattern"`
	External     map[string]string `toml:"external"`
}

type SplitLogger struct {
	requestLog *log.Logger
	errorLog   *log.Logger
	reopeners  []reopen.Reopener
}

func LoadConfig(path string) (*GatewayConfig, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	dec := toml.NewDecoder(configFile)
	config := &GatewayConfig{
		Bind:      ":8080",
		Timeout:   30000,
		Templates: templateDir,
	}
	if err := dec.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (cfg *GatewayConfig) MakeLoggers() (
	*SplitLogger,
	error,
) {
	loggers := &SplitLogger{}

	var err error
	loggers.requestLog, err = loggers.makeRotatableLogger(cfg.RequestLog)
	if err != nil {
		return nil, err
	}

	if cfg.RequestLog == cfg.ErrorLog {
		loggers.errorLog = loggers.requestLog
	} else {
		loggers.errorLog, err = loggers.makeRotatableLogger(cfg.ErrorLog)
		if err != nil {
			return nil, err
		}
	}

	return loggers, nil
}

func (logger *SplitLogger) Request(v ...interface{}) {
	logger.requestLog.Print(v...)
}

func (logger *SplitLogger) Error(v ...interface{}) {
	logger.errorLog.Print(v...)
}

func (logger *SplitLogger) Requestf(format string, v ...interface{}) {
	logger.requestLog.Printf(format, v...)
}

func (logger *SplitLogger) Errorf(format string, v ...interface{}) {
	logger.errorLog.Printf(format, v...)
}

func (logger *SplitLogger) Requestln(v ...interface{}) {
	logger.requestLog.Println(v...)
}

func (logger *SplitLogger) Errorln(v ...interface{}) {
	logger.errorLog.Println(v...)
}

func (logger *SplitLogger) Fatal(v ...interface{}) {
	logger.errorLog.Fatal(v...)
}

func (logger *SplitLogger) Reopen() error {
	for _, ro := range logger.reopeners {
		err := ro.Reopen()
		if err != nil {
			logger.Error(err)
		}
	}
	return nil
}

func (logger *SplitLogger) ServeHTTP(
	w http.ResponseWriter,
	r *http.Request,
	next http.HandlerFunc,
) {
	nrw := negroni.NewResponseWriter(w)
	start := time.Now()
	next(nrw, r)
	elapsed := time.Since(start)
	logger.requestLog.Printf("%s %s %d %d %v",
		r.Method, r.URL.Path, nrw.Status(), nrw.Size(), elapsed)
}

func (logger *SplitLogger) makeRotatableLogger(path string) (
	*log.Logger,
	error,
) {
	if path != "" {
		// Logging to file
		file, err := reopen.NewFileWriter(path)
		if err != nil {
			return nil, err
		}
		logger.reopeners = append(logger.reopeners, file)
		return log.New(file, "", log.LstdFlags), nil
	} else {
		return log.New(os.Stderr, "", log.LstdFlags), nil
	}
}
