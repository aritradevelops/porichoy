package config

import "testing"

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "8080" {
		t.Errorf("Port = %q, want 8080", cfg.Port)
	}
	if cfg.DBDSN == "" {
		t.Error("DBDSN is empty")
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	t.Setenv("PORICHOY_PORT", "9090")
	t.Setenv("PORICHOY_DB_DSN", "postgres://x:y@z:5432/w?sslmode=disable")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "9090" {
		t.Errorf("Port = %q, want 9090 (env override)", cfg.Port)
	}
	if cfg.DBDSN != "postgres://x:y@z:5432/w?sslmode=disable" {
		t.Errorf("DBDSN = %q, want env override value", cfg.DBDSN)
	}
}
