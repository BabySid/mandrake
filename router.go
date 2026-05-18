package main

import (
	"fmt"
	"log/slog"
	"net/http"
)

func BuildRouter(cfg *Config) (*http.ServeMux, error) {
	mux := http.NewServeMux()

	for _, route := range cfg.Routes {
		proxy, err := NewProxy(route)
		if err != nil {
			return nil, fmt.Errorf("route %q: %w", route.Path, err)
		}
		slog.Info("registered route", "path", route.Path, "upstream", route.Upstream)
		mux.Handle(route.Path, proxy)
	}

	return mux, nil
}
