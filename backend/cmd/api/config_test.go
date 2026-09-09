package main

import "testing"

func TestNormalizeOrigin(t *testing.T) {
	for _, tc := range []struct {
		name string
		raw  string
		want string
	}{
		{"empty", "", ""},
		{"trims whitespace", " https://application.test ", "https://application.test"},
		{"lowercases scheme and host", "HTTPS://APPLICATION.TEST", "https://application.test"},
		{"removes HTTPS default port", "https://application.test:443", "https://application.test"},
		{"removes HTTP default port", "http://application.test:80", "http://application.test"},
		{"keeps non-default port", "https://application.test:8443", "https://application.test:8443"},
		{"keeps invalid origin for fail-closed validation", ":invalid", ":invalid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := normalizeOrigin(tc.raw); got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}
