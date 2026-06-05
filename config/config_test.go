package config

import "testing"

func TestLoad_NormalizesExtraHeaderKeysWithViper(t *testing.T) {
	cfg, err := Load("../config.yaml")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Tock.ExtraHeaders == nil {
		t.Fatalf("ExtraHeaders is nil")
	}

	if _, ok := cfg.Tock.ExtraHeaders["x-toki-origin"]; !ok {
		t.Fatalf("expected key x-toki-origin in ExtraHeaders, got: %#v", cfg.Tock.ExtraHeaders)
	}

	if _, ok := cfg.Tock.ExtraHeaders["x-toki-filter"]; !ok {
		t.Fatalf("expected key x-toki-filter in ExtraHeaders, got: %#v", cfg.Tock.ExtraHeaders)
	}
}
