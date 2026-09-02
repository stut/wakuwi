package server

import "testing"

func TestHostAllowed(t *testing.T) {
	cases := []struct {
		host  string
		extra []string
		want  bool
	}{
		{"localhost", nil, true},
		{"localhost:9753", nil, true},
		{"127.0.0.1", nil, true},
		{"127.0.0.1:9753", nil, true},
		{"[::1]:9753", nil, true},
		{"::1", nil, true},
		{"evil.example.com", nil, false},
		{"evil.example.com:9753", nil, false},
		{"127.0.0.1.evil.example.com", nil, false},
		{"wakuwi.internal:9753", []string{"wakuwi.internal"}, true},
		{"Wakuwi.Internal", []string{"wakuwi.internal"}, true},
		{"other.internal", []string{"wakuwi.internal"}, false},
		{"", nil, false},
	}
	for _, c := range cases {
		if got := hostAllowed(c.host, c.extra); got != c.want {
			t.Errorf("hostAllowed(%q, %v) = %v, want %v", c.host, c.extra, got, c.want)
		}
	}
}
