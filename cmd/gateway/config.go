package main

import (
	"os"

	"github.com/pelletier/go-toml"
)

type GatewayConfig struct {
	Bind      string            `toml:"bind"`
	Root      string            `toml:"root"`
	Timeout   int64             `toml:"timeout"`
	Templates string            `toml:"templates"`
	External  map[string]string `toml:"external"`
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
