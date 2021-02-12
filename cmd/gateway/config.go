package main

import (
	"os"

	"github.com/client9/reopen"
	"github.com/pelletier/go-toml"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
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

func LoadConfig(path string) (*GatewayConfig, error) {
	configFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	dec := toml.NewDecoder(configFile)
	config := &GatewayConfig{
		Bind:    ":8080",
		Timeout: 30000,
	}
	if err := dec.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (cfg *GatewayConfig) MakeLogCore() (
	zapcore.Core,
	func(),
	error,
) {
	fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())
	consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())

	var core zapcore.Core
	var reopeners []reopen.Reopener
	var err error
	if cfg.RequestLog == cfg.ErrorLog {
		core, reopeners, err = cfg.makeUnifiedLogCore(fileEncoder, consoleEncoder)
	} else {
		core, reopeners, err = cfg.makeSplitLogCore(fileEncoder, consoleEncoder)
	}
	if err != nil {
		return nil, nil, err
	}
	reopen := func() {
		for _, re := range reopeners {
			re.Reopen()
		}
	}
	return core, reopen, nil
}

func (cfg *GatewayConfig) makeUnifiedLogCore(
	fileEncoder zapcore.Encoder,
	consoleEncoder zapcore.Encoder,
) (
	zapcore.Core,
	[]reopen.Reopener,
	error,
) {
	anyLevel := zap.LevelEnablerFunc(func(zapcore.Level) bool { return true })
	reopeners := make([]reopen.Reopener, 0)
	core, re, err := cfg.makeRotatableLogCore(
		cfg.RequestLog,
		anyLevel,
		fileEncoder,
		consoleEncoder,
	)
	if err != nil {
		return nil, nil, err
	}
	if re != nil {
		reopeners = append(reopeners, re)
	}
	return core, reopeners, nil
}

func (cfg *GatewayConfig) makeSplitLogCore(
	fileEncoder zapcore.Encoder,
	consoleEncoder zapcore.Encoder,
) (
	zapcore.Core,
	[]reopen.Reopener,
	error,
) {
	reopeners := make([]reopen.Reopener, 0)

	errorPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl >= zapcore.ErrorLevel
	})
	errorCore, re, err := cfg.makeRotatableLogCore(
		cfg.ErrorLog,
		errorPriority,
		fileEncoder,
		consoleEncoder,
	)
	if err != nil {
		return nil, nil, err
	}
	if re != nil {
		reopeners = append(reopeners, re)
	}

	requestPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl < zapcore.ErrorLevel
	})
	requestCore, re, err := cfg.makeRotatableLogCore(
		cfg.RequestLog,
		requestPriority,
		fileEncoder,
		consoleEncoder,
	)
	if err != nil {
		return nil, nil, err
	}
	if re != nil {
		reopeners = append(reopeners, re)
	}

	core := zapcore.NewTee(errorCore, requestCore)
	return core, reopeners, nil
}

func (cfg *GatewayConfig) makeRotatableLogCore(
	path string,
	enabler zapcore.LevelEnabler,
	fileEncoder zapcore.Encoder,
	consoleEncoder zapcore.Encoder,
) (
	zapcore.Core,
	reopen.Reopener,
	error,
) {
	if path != "" {
		// Logging to file
		file, err := reopen.NewFileWriter(path)
		if err != nil {
			return nil, nil, err
		}
		ws := zapcore.AddSync(file)
		core := zapcore.NewCore(fileEncoder, ws, enabler)
		return core, file, nil
	} else {
		// Logging to stderr
		ws := zapcore.Lock(os.Stderr)
		core := zapcore.NewCore(consoleEncoder, ws, enabler)
		return core, nil, nil
	}
}
