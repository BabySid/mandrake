package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestProxyHeaderModification(t *testing.T) {
	backend := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Custom") != "added" {
			t.Errorf("X-Custom = %q, want %q", r.Header.Get("X-Custom"), "added")
		}
		if r.Header.Get("X-Remove") != "" {
			t.Error("X-Remove should be deleted")
		}
		w.Header().Set("X-Backend", "original")
		w.Header().Set("X-Secret", "hidden")
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	}))
	defer backend.Close()

	route := Route{
		Upstream:      backend.URL,
		FlushInterval: -1,
		Transport:     TransportConfig{TLSSkipVerify: true},
		Request: RequestMods{
			Headers: HeaderMods{
				Set:    map[string]string{"X-Custom": "added"},
				Delete: []string{"X-Remove"},
			},
		},
		Response: ResponseMods{
			Headers: HeaderMods{
				Set:    map[string]string{"X-Proxy": "mandrake"},
				Delete: []string{"X-Secret"},
			},
		},
	}

	proxy, err := NewProxy(route)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("X-Remove", "should-go")
	w := httptest.NewRecorder()
	proxy.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
	if w.Header().Get("X-Proxy") != "mandrake" {
		t.Errorf("X-Proxy = %q, want mandrake", w.Header().Get("X-Proxy"))
	}
	if w.Header().Get("X-Secret") != "" {
		t.Error("X-Secret should be deleted from response")
	}
}

func TestProxyBodyModification(t *testing.T) {
	backend := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["model"] != "new-model" {
			t.Errorf("model = %v, want new-model", body["model"])
		}
		if _, ok := body["secret"]; ok {
			t.Error("secret should be deleted")
		}
		if body["keep"] != "this" {
			t.Errorf("keep = %v, want this", body["keep"])
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	route := Route{
		Upstream:      backend.URL,
		FlushInterval: -1,
		Transport:     TransportConfig{TLSSkipVerify: true},
		Request: RequestMods{
			Body: BodyMods{
				Set:    map[string]any{"model": "new-model"},
				Delete: []string{"secret"},
			},
		},
	}

	proxy, err := NewProxy(route)
	if err != nil {
		t.Fatal(err)
	}

	body := `{"model":"old-model","secret":"hidden","keep":"this"}`
	req := httptest.NewRequest("POST", "/test", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	proxy.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestProxyNonJSONBodyPassthrough(t *testing.T) {
	backend := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		data, _ := io.ReadAll(r.Body)
		if string(data) != "plain text body" {
			t.Errorf("body = %q, want %q", string(data), "plain text body")
		}
		w.WriteHeader(200)
	}))
	defer backend.Close()

	route := Route{
		Upstream:      backend.URL,
		FlushInterval: -1,
		Transport:     TransportConfig{TLSSkipVerify: true},
		Request: RequestMods{
			Body: BodyMods{
				Set: map[string]any{"model": "should-not-apply"},
			},
		},
	}

	proxy, err := NewProxy(route)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", "/test", strings.NewReader("plain text body"))
	req.Header.Set("Content-Type", "text/plain")
	w := httptest.NewRecorder()
	proxy.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("status = %d, want 200", w.Code)
	}
}

func TestProxySSEStreaming(t *testing.T) {
	backend := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected flusher")
		}
		for _, event := range []string{"data: hello\n\n", "data: world\n\n"} {
			w.Write([]byte(event))
			flusher.Flush()
		}
	}))
	defer backend.Close()

	route := Route{
		Upstream:      backend.URL,
		FlushInterval: -1,
		Transport:     TransportConfig{TLSSkipVerify: true},
	}

	proxy, err := NewProxy(route)
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", "/events", nil)
	w := httptest.NewRecorder()
	proxy.ServeHTTP(w, req)

	body := w.Body.String()
	if !strings.Contains(body, "data: hello") || !strings.Contains(body, "data: world") {
		t.Errorf("SSE body = %q, expected hello and world events", body)
	}
}
