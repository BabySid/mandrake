package main

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	raw := `{
		"server": {
			"listen": ":9090",
			"read_timeout": "30s",
			"write_timeout": "60s"
		},
		"routes": [
			{
				"path": "/api/",
				"upstream": "https://backend.example.com",
				"flush_interval": -1,
				"transport": {
					"tls_skip_verify": true,
					"read_timeout": "120s",
					"write_timeout": "120s"
				},
				"request": {
					"headers": {
						"set": {"Host": "backend.example.com"},
						"delete": ["X-Debug"]
					},
					"body": {
						"set": {"model": "claude-sonnet-4-20250514"},
						"delete": ["internal"]
					}
				},
				"response": {
					"headers": {
						"set": {"X-Proxy": "mandrake"},
						"delete": ["Server"]
					}
				}
			}
		],
		"log": {
			"level": "debug"
		}
	}`

	f, err := os.CreateTemp("", "mandrake-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(raw)
	f.Close()

	cfg, err := LoadConfig(f.Name())
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.Server.Listen != ":9090" {
		t.Errorf("listen = %q, want %q", cfg.Server.Listen, ":9090")
	}
	if cfg.Server.ReadTimeout.String() != "30s" {
		t.Errorf("read_timeout = %v, want 30s", cfg.Server.ReadTimeout)
	}
	if len(cfg.Routes) != 1 {
		t.Fatalf("routes count = %d, want 1", len(cfg.Routes))
	}

	r := cfg.Routes[0]
	if r.Path != "/api/" {
		t.Errorf("path = %q, want %q", r.Path, "/api/")
	}
	if r.Upstream != "https://backend.example.com" {
		t.Errorf("upstream = %q", r.Upstream)
	}
	if r.FlushInterval != -1 {
		t.Errorf("flush_interval = %d, want -1", r.FlushInterval)
	}
	if r.Transport.TLSSkipVerify != true {
		t.Error("tls_skip_verify should be true")
	}
	if r.Request.Headers.Set["Host"] != "backend.example.com" {
		t.Errorf("request header set Host = %q", r.Request.Headers.Set["Host"])
	}
	if len(r.Request.Headers.Delete) != 1 || r.Request.Headers.Delete[0] != "X-Debug" {
		t.Errorf("request header delete = %v", r.Request.Headers.Delete)
	}
	if r.Request.Body.Set["model"] != "claude-sonnet-4-20250514" {
		t.Errorf("request body set model = %v", r.Request.Body.Set["model"])
	}
	if r.Response.Headers.Set["X-Proxy"] != "mandrake" {
		t.Errorf("response header set X-Proxy = %q", r.Response.Headers.Set["X-Proxy"])
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	raw := `{
		"routes": [
			{
				"path": "/",
				"upstream": "https://example.com"
			}
		]
	}`

	f, err := os.CreateTemp("", "mandrake-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString(raw)
	f.Close()

	cfg, err := LoadConfig(f.Name())
	if err != nil {
		t.Fatalf("LoadConfig: %v", err)
	}

	if cfg.Server.Listen != ":8080" {
		t.Errorf("default listen = %q, want %q", cfg.Server.Listen, ":8080")
	}
}

func TestLoadConfigInvalid(t *testing.T) {
	f, err := os.CreateTemp("", "mandrake-*.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(f.Name())
	f.WriteString("not json")
	f.Close()

	_, err = LoadConfig(f.Name())
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}
