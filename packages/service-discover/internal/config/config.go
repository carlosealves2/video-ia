package config

import (
	"errors"
	"os"
	"strconv"
)

type Config struct {
	Port     int
	LogLevel string
	GinMode  string
}

type Builder struct {
	config *Config
	errors []error
}

func NewBuilder() *Builder {
	return &Builder{
		config: &Config{
			Port:     8080,
			LogLevel: "info",
			GinMode:  "release",
		},
		errors: []error{},
	}
}

func (b *Builder) WithEnv() *Builder {
	if port := os.Getenv("PORT"); port != "" {
		p, err := strconv.Atoi(port)
		if err != nil {
			b.errors = append(b.errors, errors.New("PORT must be a valid integer"))
		} else {
			b.config.Port = p
		}
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		b.config.LogLevel = logLevel
	}

	if ginMode := os.Getenv("GIN_MODE"); ginMode != "" {
		b.config.GinMode = ginMode
	}

	return b
}

func (b *Builder) Validate() *Builder {
	if b.config.Port <= 0 || b.config.Port > 65535 {
		b.errors = append(b.errors, errors.New("PORT must be between 1 and 65535"))
	}

	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}
	if !validLogLevels[b.config.LogLevel] {
		b.errors = append(b.errors, errors.New("LOG_LEVEL must be one of: debug, info, warn, error"))
	}

	validGinModes := map[string]bool{
		"debug":   true,
		"release": true,
		"test":    true,
	}
	if !validGinModes[b.config.GinMode] {
		b.errors = append(b.errors, errors.New("GIN_MODE must be one of: debug, release, test"))
	}

	return b
}

func (b *Builder) Build() (*Config, error) {
	if len(b.errors) > 0 {
		return nil, errors.Join(b.errors...)
	}
	return b.config, nil
}
