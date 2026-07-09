package server

import "testing"

func TestParseSemver(t *testing.T) {
	for _, tc := range []struct {
		in string
		ok bool
	}{
		{"0.5.0", true},
		{"v0.5.0", true},
		{"0.5.0-dev", true},
		{"dev", false},
		{"1.2", false},
		{"", false},
	} {
		if _, ok := parseSemver(tc.in); ok != tc.ok {
			t.Errorf("parseSemver(%q) ok = %v, want %v", tc.in, ok, tc.ok)
		}
	}
}

func TestSemverLess(t *testing.T) {
	for _, tc := range []struct {
		a, b string
		less bool
	}{
		{"0.4.0", "0.5.0", true},
		{"0.5.0", "0.4.0", false},
		{"0.5.0", "0.5.0", false},
		{"0.5.0-dev", "0.5.0", true},
		{"0.5.0", "0.5.0-dev", false},
		{"0.5.0-dev", "0.4.0", false},
		{"1.9.0", "1.10.0", true},
		{"0.5.0", "1.0.0", true},
	} {
		a, _ := parseSemver(tc.a)
		b, _ := parseSemver(tc.b)
		if got := a.less(b); got != tc.less {
			t.Errorf("%q.less(%q) = %v, want %v", tc.a, tc.b, got, tc.less)
		}
	}
}
