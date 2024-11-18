package main

import (
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"
)

// -ldflags "-X main.version=<value>"
var version = "local"

func main() {
	port := "8080"
	if p := os.Getenv("PORT"); p != "" {
		port = p
	}

	l := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug}))
	c := CorsAnywhere{
		Log:            l,
		RequireHeaders: []string{"Origin"},
		RemoveHeaders:  []string{"Set-Cookie", "Set-Cookie2"},
		MaxAge:         86400,
	}

	l.Info("listening", "port", port, "version", version)
	if err := http.ListenAndServe(":"+port, c.Proxy()); err != nil {
		slog.Error("listening error", "error", err)
		os.Exit(1)
	}
}

type CorsAnywhere struct {
	Log            *slog.Logger
	RequireHeaders []string
	RemoveHeaders  []string
	MaxAge         uint
}

func (c CorsAnywhere) Proxy() http.Handler {
	proxy := httputil.ReverseProxy{
		Director:       c.director,
		ModifyResponse: c.modifyResponse,
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			DialContext: (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout: 10 * time.Second,
		},
		ErrorLog: slog.NewLogLogger(c.Log.Handler(), slog.LevelError),
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c.Log.InfoContext(r.Context(), "handling request", "method", r.Method, "url", r.RequestURI)

		if r.RequestURI == "/" {
			c.handleHome(w, r)
			return
		}

		trimmedURI := trimURL(r.RequestURI)
		u, err := url.Parse(trimmedURI)
		if err != nil || !slices.Contains([]string{"http", "https"}, u.Scheme) || u.Host == "" {
			c.handleError(w, r, "unprocessable destination url: "+strings.TrimPrefix(r.RequestURI, "/"))
			return
		}

		if r.Method == http.MethodOptions {
			handlePreflight(w, r, int(c.MaxAge))
			return
		}

		for _, h := range c.RequireHeaders {
			if r.Header.Get(h) == "" {
				c.handleError(w, r, "missing required header: "+h)
				return
			}
		}

		proxy.ServeHTTP(w, r)
	})
}

func (c CorsAnywhere) director(r *http.Request) {
	u, err := url.Parse(trimURL(r.RequestURI))
	if err != nil {
		c.Log.ErrorContext(r.Context(), "director: failed to parse url", "url", r.RequestURI, "error", err)
		return
	}
	c.Log.DebugContext(r.Context(), "destination url", "url", u)

	r.URL.Scheme = u.Scheme
	r.URL.Host = u.Host
	r.URL.Path = u.Path
	r.Host = u.Host

	for _, h := range c.RemoveHeaders {
		r.Header.Del(h)
	}
}

func (c CorsAnywhere) modifyResponse(r *http.Response) error {
	r.Header.Set("Access-Control-Allow-Origin", "*")
	r.Header.Set("Access-Control-Max-Age", strconv.Itoa(int(c.MaxAge)))
	return nil
}

func (c CorsAnywhere) handleHome(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("corsanywhere")); err != nil {
		c.Log.ErrorContext(r.Context(), "error writing error response", "error", err)
	}
}

func (c CorsAnywhere) handleError(w http.ResponseWriter, r *http.Request, message string) {
	w.WriteHeader(http.StatusBadRequest)
	if _, err := w.Write([]byte(message)); err != nil {
		c.Log.ErrorContext(r.Context(), "error writing error response", "error", err)
	}
}

func handlePreflight(w http.ResponseWriter, r *http.Request, maxAge int) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Max-Age", strconv.Itoa(maxAge))
	if r.Header.Get("Access-Control-Request-Method") != "" {
		w.Header().Set("Access-Control-Allow-Methods", r.Header.Get("Access-Control-Request-Method"))
	}
	if r.Header.Get("Access-Control-Request-Headers") != "" {
		w.Header().Set("Access-Control-Request-Headers", r.Header.Get("Access-Control-Request-Headers"))
	}
	w.WriteHeader(http.StatusOK)
}

func trimURL(uri string) string {
	uri = strings.TrimLeft(uri, "/")
	u, err := url.Parse(uri)
	if err != nil {
		return uri
	}
	if u.Scheme == "" {
		return "https://" + uri
	}
	return uri
}
