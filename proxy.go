package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"
)

func NewProxy(route Route) (http.Handler, error) {
	upstream, err := url.Parse(route.Upstream)
	if err != nil {
		return nil, fmt.Errorf("parse upstream %q: %w", route.Upstream, err)
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: route.Transport.TLSSkipVerify,
		},
	}
	if route.Transport.ReadTimeout.Duration > 0 || route.Transport.WriteTimeout.Duration > 0 {
		timeout := route.Transport.ReadTimeout.Duration
		if route.Transport.WriteTimeout.Duration > timeout {
			timeout = route.Transport.WriteTimeout.Duration
		}
		transport.ResponseHeaderTimeout = timeout
	}

	hasBodyMods := len(route.Request.Body.Set) > 0 || len(route.Request.Body.Delete) > 0 || len(route.Request.Body.Prepend) > 0 || len(route.Request.Body.Append) > 0
	bodyMods := route.Request.Body

	rt := &bodyModifyingTransport{
		base:        transport,
		hasBodyMods: hasBodyMods,
		bodyMods:    bodyMods,
	}

	proxy := &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.URL.Scheme = upstream.Scheme
			req.URL.Host = upstream.Host
			req.Host = upstream.Host

			for key, value := range route.Request.Headers.Set {
				if strings.EqualFold(key, "host") {
					req.Host = value
				}
				req.Header.Set(key, value)
			}
			for _, key := range route.Request.Headers.Delete {
				req.Header.Del(key)
			}

			slog.Debug("request headers after modification", "method", req.Method, "path", req.URL.Path, "headers", req.Header)
		},
		Transport: rt,
		ModifyResponse: func(resp *http.Response) error {
			for key, value := range route.Response.Headers.Set {
				resp.Header.Set(key, value)
			}
			for _, key := range route.Response.Headers.Delete {
				resp.Header.Del(key)
			}
			return nil
		},
		FlushInterval: time.Duration(route.FlushInterval) * time.Millisecond,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			slog.Error("proxy error", "upstream", route.Upstream, "path", r.URL.Path, "error", err)
			w.WriteHeader(http.StatusBadGateway)
		},
	}

	return proxy, nil
}

type bodyModifyingTransport struct {
	base        http.RoundTripper
	hasBodyMods bool
	bodyMods    BodyMods
}

func (t *bodyModifyingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.hasBodyMods && req.Body != nil && isJSONContent(req.Header.Get("Content-Type")) {
		data, err := io.ReadAll(req.Body)
		req.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("read request body: %w", err)
		}

		var obj map[string]any
		if err := json.Unmarshal(data, &obj); err != nil {
			slog.Warn("body is not valid JSON, skipping modification", "error", err)
			req.Body = io.NopCloser(bytes.NewReader(data))
			req.ContentLength = int64(len(data))
		} else {
			ModifyBody(obj, t.bodyMods)
			modified, err := json.Marshal(obj)
			if err != nil {
				return nil, fmt.Errorf("marshal modified body: %w", err)
			}
			var pretty bytes.Buffer
			json.Indent(&pretty, modified, "", "  ")
			slog.Debug("request body after modification", "body", pretty.String())
			req.Body = io.NopCloser(bytes.NewReader(modified))
			req.ContentLength = int64(len(modified))
		}
	}
	return t.base.RoundTrip(req)
}

func isJSONContent(ct string) bool {
	return strings.HasPrefix(ct, "application/json")
}
