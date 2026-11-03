// Package config reads runtime settings from environment variables.
package config

import (
	"errors"

	"github.com/xkovacikm2/ssd-virtual-battery/internal/ssd"
)

type Config struct {
	DatabaseURL string
	Addr        string
	SSD         ssd.Config
}

func Load(getenv func(string) string) (Config, error) {
	cfg := Config{
		DatabaseURL: getenv("DATABASE_URL"),
		Addr:        ":" + or(getenv("PORT"), "3000"),
		SSD: ssd.Config{
			BaseURL:           getenv("IMS_SSD_BASE_URL"),
			Username:          getenv("IMS_SSD_USERNAME"),
			Password:          getenv("IMS_SSD_PASSWORD"),
			PointOfDeliveryID: getenv("IMS_SSD_POINT_OF_DELIVERY_ID"),
		},
	}
	if cfg.DatabaseURL == "" {
		return Config{}, errors.New("DATABASE_URL environment variable is required")
	}
	return cfg, nil
}

func or(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}
