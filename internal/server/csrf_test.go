package server

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCsrfOK(t *testing.T) {
	cases := []struct {
		name    string
		headers map[string]string
		body    string
		want    bool
	}{
		{"no browser headers, no body", nil, "", true},
		{"same-origin fetch with json", map[string]string{
			"Sec-Fetch-Site": "same-origin",
			"Origin":         "http://localhost:9753",
			"Content-Type":   "application/json",
		}, "{}", true},
		{"json with charset", map[string]string{
			"Content-Type": "application/json; charset=utf-8",
		}, "{}", true},
		{"cross-site sec-fetch-site", map[string]string{
			"Sec-Fetch-Site": "cross-site",
			"Content-Type":   "application/json",
		}, "{}", false},
		{"cross-origin simple request", map[string]string{
			"Origin":       "http://evil.example.com",
			"Content-Type": "text/plain",
		}, "{}", false},
		{"null origin", map[string]string{
			"Origin":       "null",
			"Content-Type": "application/json",
		}, "{}", false},
		{"body without json content type", map[string]string{
			"Content-Type": "text/plain",
		}, "{}", false},
		{"body without any content type", nil, "{}", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			r := httptest.NewRequest("POST", "/api/processes", strings.NewReader(c.body))
			r.Header.Del("Content-Type")
			for k, v := range c.headers {
				r.Header.Set(k, v)
			}
			if got := csrfOK(r); got != c.want {
				t.Errorf("csrfOK() = %v, want %v", got, c.want)
			}
		})
	}
}
