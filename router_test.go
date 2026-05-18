package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildRouterSingleRoute(t *testing.T) {
	backend := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer backend.Close()

	cfg := &Config{
		Routes: []Route{
			{
				Path:          "/",
				Upstream:      backend.URL,
				FlushInterval: -1,
				Transport:     TransportConfig{TLSSkipVerify: true},
			},
		},
	}

	mux, err := BuildRouter(cfg)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/anything", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestBuildRouterMultipleRoutes(t *testing.T) {
	backend1 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend1"))
	}))
	defer backend1.Close()

	backend2 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("backend2"))
	}))
	defer backend2.Close()

	cfg := &Config{
		Routes: []Route{
			{
				Path:      "/api/",
				Upstream:  backend1.URL,
				Transport: TransportConfig{TLSSkipVerify: true},
			},
			{
				Path:      "/",
				Upstream:  backend2.URL,
				Transport: TransportConfig{TLSSkipVerify: true},
			},
		},
	}

	mux, err := BuildRouter(cfg)
	if err != nil {
		t.Fatal(err)
	}

	// /api/ should go to backend1
	req1 := httptest.NewRequest("GET", "/api/test", nil)
	w1 := httptest.NewRecorder()
	mux.ServeHTTP(w1, req1)
	if w1.Body.String() != "backend1" {
		t.Errorf("/api/test body = %q, want backend1", w1.Body.String())
	}

	// / should go to backend2
	req2 := httptest.NewRequest("GET", "/other", nil)
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, req2)
	if w2.Body.String() != "backend2" {
		t.Errorf("/other body = %q, want backend2", w2.Body.String())
	}
}
