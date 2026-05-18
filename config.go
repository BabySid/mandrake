package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	dur, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	d.Duration = dur
	return nil
}

func (d Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(d.Duration.String())
}

type Config struct {
	Server ServerConfig `json:"server"`
	Routes []Route      `json:"routes"`
	Log    LogConfig    `json:"log"`
}

type ServerConfig struct {
	Listen       string   `json:"listen"`
	ReadTimeout  Duration `json:"read_timeout"`
	WriteTimeout Duration `json:"write_timeout"`
}

type Route struct {
	Path          string          `json:"path"`
	Upstream      string          `json:"upstream"`
	FlushInterval int             `json:"flush_interval"`
	Transport     TransportConfig `json:"transport"`
	Request       RequestMods     `json:"request"`
	Response      ResponseMods    `json:"response"`
}

type TransportConfig struct {
	TLSSkipVerify bool     `json:"tls_skip_verify"`
	ReadTimeout   Duration `json:"read_timeout"`
	WriteTimeout  Duration `json:"write_timeout"`
}

type RequestMods struct {
	Headers HeaderMods `json:"headers"`
	Body    BodyMods   `json:"body"`
}

type ResponseMods struct {
	Headers HeaderMods `json:"headers"`
}

type HeaderMods struct {
	Set    map[string]string `json:"set"`
	Delete []string          `json:"delete"`
}

type BodyMods struct {
	Set    map[string]any `json:"set"`
	Delete []string       `json:"delete"`
}

type LogConfig struct {
	Level string `json:"level"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.Server.Listen == "" {
		cfg.Server.Listen = ":8080"
	}

	return &cfg, nil
}
