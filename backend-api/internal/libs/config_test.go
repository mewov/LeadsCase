package libs

import (
	"strings"
	"testing"
	"time"
)

func TestCreateConfigSuccess(t *testing.T) {
	setValidConfigEnv(t)
	t.Setenv("SERVER_ADMIN_LOGIN", " Admin ")

	config, err := CreateConfig()
	if err != nil {
		t.Fatalf("CreateConfig() error = %v", err)
	}

	if config.ServerAddr != "0.0.0.0:80" {
		t.Fatalf("ServerAddr = %q", config.ServerAddr)
	}
	if config.ServerReadTimeout != 15*time.Second {
		t.Fatalf("ServerReadTimeout = %s", config.ServerReadTimeout)
	}
	if config.ServerWriteTimeout != 20*time.Second {
		t.Fatalf("ServerWriteTimeout = %s", config.ServerWriteTimeout)
	}
	if config.ServerAdminLogin != "admin" {
		t.Fatalf("ServerAdminLogin = %q", config.ServerAdminLogin)
	}
	if config.JWTSecret != "jwt-secret" {
		t.Fatalf("JWTSecret = %q", config.JWTSecret)
	}
}

func TestCreateConfigRequiresEnv(t *testing.T) {
	setValidConfigEnv(t)
	t.Setenv("JWT_SECRET", " ")

	_, err := CreateConfig()
	if err == nil {
		t.Fatal("CreateConfig() error = nil")
	}
	if !strings.Contains(err.Error(), "JWT_SECRET is required") {
		t.Fatalf("CreateConfig() error = %q", err)
	}
}

func TestCreateConfigRequiresPositiveTimeout(t *testing.T) {
	setValidConfigEnv(t)
	t.Setenv("SERVER_READ_TIMEOUT_SECOND", "0")

	_, err := CreateConfig()
	if err == nil {
		t.Fatal("CreateConfig() error = nil")
	}
	if !strings.Contains(err.Error(), "SERVER_READ_TIMEOUT_SECOND must be greater than 0") {
		t.Fatalf("CreateConfig() error = %q", err)
	}
}

func setValidConfigEnv(t *testing.T) {
	t.Helper()

	values := map[string]string{
		"SERVER_ADDR":                 "0.0.0.0:80",
		"SERVER_READ_TIMEOUT_SECOND":  "15",
		"SERVER_WRITE_TIMEOUT_SECOND": "20",
		"SERVER_ADMIN_LOGIN":          "admin",
		"SERVER_ADMIN_PASSWORD":       "admin-password",
		"JWT_SECRET":                  "jwt-secret",
		"SERVER_LOG_PATH":             "./logs/server.log",
		"SERVER_HTTP_LOG_PATH":        "./logs/server_http.log",
		"SERVER_POSTGRES_LOG_PATH":    "./logs/server_postgres.log",
		"POSTGRES_ADDR":               "postgres:5432",
		"POSTGRES_USER":               "admin",
		"POSTGRES_PASSWORD":           "postgres-password",
		"POSTGRES_DB":                 "demo",
	}

	for key, value := range values {
		t.Setenv(key, value)
	}
}
