package config

import "testing"

func TestLoad_PreservesExtraHeaderKeys(t *testing.T) {
	cfg, err := Load("../config.yaml")
	if err != nil {
		t.Fatalf("Load() error: %v", err)
	}

	if cfg.Tock.ExtraHeaders == nil {
		t.Fatalf("ExtraHeaders is nil")
	}

	if _, ok := cfg.Tock.ExtraHeaders["X-Toki-Origin"]; !ok {
		t.Fatalf("expected key X-Toki-Origin in ExtraHeaders, got: %#v", cfg.Tock.ExtraHeaders)
	}

	if _, ok := cfg.Tock.ExtraHeaders["X-Toki-Filter"]; !ok {
		t.Fatalf("expected key X-Toki-Filter in ExtraHeaders, got: %#v", cfg.Tock.ExtraHeaders)
	}
}
