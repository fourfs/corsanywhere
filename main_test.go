package main

import "testing"

func Test_trimURL(t *testing.T) {
	tests := []struct {
		uri  string
		want string
	}{
		{"https://example.com", "https://example.com"},
		{"http://example.com", "http://example.com"},
		{"https:/example.com", "https:/example.com"},
		{"a://example.com", "a://example.com"},
		{"example.com", "https://example.com"},
		{"////", "https://"},
		{"///example.com", "https://example.com"},
		{"https:///example.com", "https:///example.com"},
	}
	for _, tt := range tests {
		t.Run(tt.uri, func(t *testing.T) {
			if got := trimURL(tt.uri); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
