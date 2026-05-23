package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	Server       ServerConfig           `json:"server"`
	DefaultModel string                 `json:"default_model"`
	Models       map[string]ModelConfig `json:"models"`
}

type ServerConfig struct {
	Port       int `json:"port"`
	Workers    int `json:"workers"`
	CallbackWorkers int `json:"callback_workers"`
}

// ModelConfig is a union of every adapter's accepted settings; unused fields
// for a given adapter are simply ignored.
type ModelConfig struct {
	Adapter     string `json:"adapter"`
	Binary      string `json:"binary,omitempty"`
	ModelFile   string `json:"model_file,omitempty"`
	Threads     int    `json:"threads,omitempty"`
	Model       string `json:"model,omitempty"`
	ComputeType string `json:"compute_type,omitempty"`
	Device      string `json:"device,omitempty"`
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	cfg.applyDefaults()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *Config) applyDefaults() {
	if c.Server.Port == 0 {
		c.Server.Port = 8888
	}
	if c.Server.Workers == 0 {
		c.Server.Workers = 2
	}
	if c.Server.CallbackWorkers == 0 {
		c.Server.CallbackWorkers = 2
	}
}

func (c *Config) validate() error {
	if len(c.Models) == 0 {
		return fmt.Errorf("config: at least one model must be configured")
	}
	if c.DefaultModel == "" {
		return fmt.Errorf("config: default_model is required")
	}
	if _, ok := c.Models[c.DefaultModel]; !ok {
		return fmt.Errorf("config: default_model %q not in models", c.DefaultModel)
	}
	return nil
}
