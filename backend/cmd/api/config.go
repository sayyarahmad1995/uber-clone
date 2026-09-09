package main

import (
	"net"
	"net/url"
	"os"
	"strings"
)

type config struct {
	Port                string
	DatabaseURL         string
	AuthProvider        string
	IdentitySource      string
	KratosPublicURL     string
	KratosAdminURL      string
	AdminReviewUsername string
	AdminReviewPassword string
	AdminReviewOrigin   string
}

func loadConfig() config {
	return config{
		Port:                getenv("APP_PORT", "8080"),
		DatabaseURL:         os.Getenv("DATABASE_URL"),
		AuthProvider:        getenv("AUTH_PROVIDER", "kratos"),
		IdentitySource:      getenv("AUTH_IDENTITY_SOURCE", "primary-identity-v1"),
		KratosPublicURL:     getenv("KRATOS_PUBLIC_URL", "http://kratos:4433"),
		KratosAdminURL:      getenv("KRATOS_ADMIN_URL", "http://kratos:4434"),
		AdminReviewUsername: os.Getenv("ADMIN_REVIEW_USERNAME"),
		AdminReviewPassword: os.Getenv("ADMIN_REVIEW_PASSWORD"),
		AdminReviewOrigin:   normalizeOrigin(os.Getenv("ADMIN_REVIEW_ORIGIN")),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func normalizeOrigin(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return trimmed
	}
	scheme := strings.ToLower(parsed.Scheme)
	if (scheme != "http" && scheme != "https") || parsed.Host == "" || parsed.User != nil || parsed.Path != "" || parsed.RawQuery != "" || parsed.Fragment != "" {
		return trimmed
	}
	host := strings.ToLower(parsed.Host)
	if hostname, port, err := net.SplitHostPort(parsed.Host); err == nil {
		if (scheme == "https" && port == "443") || (scheme == "http" && port == "80") {
			host = strings.ToLower(hostname)
			if strings.Contains(hostname, ":") {
				host = "[" + host + "]"
			}
		}
	}
	return scheme + "://" + host
}
