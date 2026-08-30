package main

import "os"

type config struct {
	Port            string
	DatabaseURL     string
	AuthProvider    string
	IdentitySource  string
	KratosPublicURL string
	KratosAdminURL  string
}

func loadConfig() config {
	return config{
		Port:            getenv("APP_PORT", "8080"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		AuthProvider:    getenv("AUTH_PROVIDER", "kratos"),
		IdentitySource:  getenv("AUTH_IDENTITY_SOURCE", "primary-identity-v1"),
		KratosPublicURL: getenv("KRATOS_PUBLIC_URL", "http://kratos:4433"),
		KratosAdminURL:  getenv("KRATOS_ADMIN_URL", "http://kratos:4434"),
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
