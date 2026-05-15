package config

import "os"

type Config struct {
	WebAppAddr string
}

var Default Config

func init() {
	cfg := Config{}
	if os.Getenv("WEBAPP_ADDR") != "" {
		cfg.WebAppAddr = os.Getenv("WEBAPP_ADDR")
	}
	Default = cfg
}
