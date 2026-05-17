package config

import (
	"os"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()
	if cfg == nil {
		t.Fatal("Default() returned nil")
	}
	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %s", cfg.Server.Host)
	}
	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}
	if cfg.JWT.Secret != "change-me-to-a-random-secret" {
		t.Errorf("unexpected JWT secret")
	}
	if cfg.JWT.TTL != 86400 {
		t.Errorf("expected TTL 86400, got %d", cfg.JWT.TTL)
	}
	if cfg.Panel.AdminUsername != "admin" {
		t.Errorf("expected admin username 'admin', got %s", cfg.Panel.AdminUsername)
	}
	if cfg.Panel.Domain != "example.com" {
		t.Errorf("expected domain 'example.com', got %s", cfg.Panel.Domain)
	}
}

func TestLoad(t *testing.T) {
	yamlContent := `
server:
  host: "127.0.0.1"
  port: 9090
database:
  dsn: "postgres://test:test@localhost:5432/test"
jwt:
  secret: "test-secret"
  ttl: 3600
panel:
  admin_username: "testadmin"
  admin_password: "testpass"
  domain: "test.example.com"
`
	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlContent)); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	tmpFile.Close()

	cfg, err := Load(tmpFile.Name())
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("host: got %s, want 127.0.0.1", cfg.Server.Host)
	}
	if cfg.Server.Port != 9090 {
		t.Errorf("port: got %d, want 9090", cfg.Server.Port)
	}
	if cfg.Database.DSN != "postgres://test:test@localhost:5432/test" {
		t.Errorf("dsn: got %s, want postgres://test:test@localhost:5432/test", cfg.Database.DSN)
	}
	if cfg.JWT.Secret != "test-secret" {
		t.Errorf("secret: got %s, want test-secret", cfg.JWT.Secret)
	}
	if cfg.Panel.Domain != "test.example.com" {
		t.Errorf("domain: got %s, want test.example.com", cfg.Panel.Domain)
	}
}

func TestLoad_fileNotFound(t *testing.T) {
	_, err := Load("/nonexistent/path/config.yaml")
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}
