// corsanywhere - get past cors restricitons.
// Copyright (C) 2024 fourfs
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

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
	requireHeaders := []string{}
	if rh := os.Getenv("REQUIRE_HEADERS"); rh != "" {
		requireHeaders = strings.Split(rh, ",")
	}
	removeHeaders := []string{"Set-Cookie", "Set-Cookie2"}
	if rh := os.Getenv("REMOVE_HEADERS"); rh != "" {
		removeHeaders = strings.Split(rh, ",")
	}
	logLevel := slog.LevelInfo
	if ll := os.Getenv("LOG_LEVEL"); ll != "" {
		li, err := strconv.Atoi(ll)
		if err != nil {
			logLevel = slog.Level(li)
		}
	}

	c := CorsAnywhere{
		Log:            slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})),
		RequireHeaders: requireHeaders,
		RemoveHeaders:  removeHeaders,
		MaxAge:         86400,
	}

	c.Log.Info("listening", "port", port, "version", version)
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
		Rewrite:        c.rewrite,
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
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			c.handleError(w, r, err.Error())
		},
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

func (c CorsAnywhere) rewrite(r *httputil.ProxyRequest) {
	u, err := url.Parse(trimURL(r.In.RequestURI))
	if err != nil {
		c.Log.ErrorContext(r.In.Context(), "director: failed to parse url", "url", r.In.RequestURI, "error", err)
		return
	}
	c.Log.DebugContext(r.In.Context(), "destination url", "url", u)

	r.Out.URL.Scheme = u.Scheme
	r.Out.URL.Host = u.Host
	r.Out.URL.Path = u.Path
	r.Out.Host = u.Host

	for _, h := range c.RemoveHeaders {
		r.Out.Header.Del(h)
	}

	if origin := r.In.Header.Get("X-Set-Origin"); origin != "" {
		c.Log.DebugContext(r.In.Context(), "setting origin", "origin", origin)
		r.Out.Header.Del("X-Set-Origin")
		r.Out.Header.Set("Origin", origin)
		r.Out.Header.Set("Referer", origin+"/")
	}
}

func (c CorsAnywhere) modifyResponse(r *http.Response) error {
	r.Header.Set("Access-Control-Allow-Origin", "*")
	r.Header.Set("Access-Control-Max-Age", strconv.Itoa(int(c.MaxAge)))
	return nil
}

func (c CorsAnywhere) handleHome(w http.ResponseWriter, r *http.Request) {
	if _, err := w.Write([]byte("corsanywhere")); err != nil {
		c.Log.ErrorContext(r.Context(), "error writing home response", "error", err)
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
		w.Header().Set("Access-Control-Allow-Headers", r.Header.Get("Access-Control-Request-Headers"))
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
