package config

import "testing"

func TestFromEnvDefaults(t *testing.T) {
	cfg := FromEnv()
	if cfg.Port != 8083 {
		t.Errorf("expected default port 8083, got %d", cfg.Port)
	}
	if cfg.LogLevel != "info" {
		t.Errorf("expected default log level info, got %q", cfg.LogLevel)
	}
	if cfg.FailReadiness {
		t.Error("expected FailReadiness to default to false")
	}
}

func TestFromEnvOverrides(t *testing.T) {
	t.Setenv("PORT", "9999")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("FAIL_READINESS", "true")

	cfg := FromEnv()
	if cfg.Port != 9999 {
		t.Errorf("expected port 9999, got %d", cfg.Port)
	}
	if cfg.LogLevel != "debug" {
		t.Errorf("expected log level debug, got %q", cfg.LogLevel)
	}
	if !cfg.FailReadiness {
		t.Error("expected FailReadiness true")
	}
}
