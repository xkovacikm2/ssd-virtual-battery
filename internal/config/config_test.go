package config

import (
	"strings"
	"testing"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load(env(map[string]string{"DATABASE_URL": "postgres://x"}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DatabaseURL != "postgres://x" || cfg.Addr != ":3000" {
		t.Fatalf("%+v", cfg)
	}
	if cfg.SSD.BaseURL != "" {
		t.Fatalf("SSD base URL should be left to the client default, got %q", cfg.SSD.BaseURL)
	}
}

func TestLoadReadsEverything(t *testing.T) {
	cfg, err := Load(env(map[string]string{
		"DATABASE_URL":                 "postgres://x",
		"PORT":                         "8080",
		"IMS_SSD_BASE_URL":             "http://fake",
		"IMS_SSD_USERNAME":             "u",
		"IMS_SSD_PASSWORD":             "p",
		"IMS_SSD_POINT_OF_DELIVERY_ID": "42",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Addr != ":8080" || cfg.SSD.BaseURL != "http://fake" || cfg.SSD.Username != "u" ||
		cfg.SSD.Password != "p" || cfg.SSD.PointOfDeliveryID != "42" {
		t.Fatalf("%+v", cfg)
	}
}

func TestLoadRequiresDatabaseURL(t *testing.T) {
	_, err := Load(env(map[string]string{}))
	if err == nil || !strings.Contains(err.Error(), "DATABASE_URL") {
		t.Fatalf("err = %v", err)
	}
}

func TestSSDCredentialsAreOptionalAtStartup(t *testing.T) {
	if _, err := Load(env(map[string]string{"DATABASE_URL": "postgres://x"})); err != nil {
		t.Fatalf("dashboard must start without SSD credentials: %v", err)
	}
}
