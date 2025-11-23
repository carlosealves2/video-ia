package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBuilder(t *testing.T) {
	builder := NewBuilder()
	assert.NotNil(t, builder)
	assert.NotNil(t, builder.config)
	assert.Equal(t, 8080, builder.config.Port)
	assert.Equal(t, "info", builder.config.LogLevel)
	assert.Equal(t, "release", builder.config.GinMode)
}

func TestWithEnvDefaults(t *testing.T) {
	// Clear any existing env vars
	_ = os.Unsetenv("PORT")
	_ = os.Unsetenv("LOG_LEVEL")
	_ = os.Unsetenv("GIN_MODE")

	cfg, err := NewBuilder().WithEnv().Validate().Build()

	require.NoError(t, err)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, "info", cfg.LogLevel)
	assert.Equal(t, "release", cfg.GinMode)
}

func TestWithEnvCustomValues(t *testing.T) {
	_ = os.Setenv("PORT", "3000")
	_ = os.Setenv("LOG_LEVEL", "debug")
	_ = os.Setenv("GIN_MODE", "debug")
	defer func() {
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("LOG_LEVEL")
		_ = os.Unsetenv("GIN_MODE")
	}()

	cfg, err := NewBuilder().WithEnv().Validate().Build()

	require.NoError(t, err)
	assert.Equal(t, 3000, cfg.Port)
	assert.Equal(t, "debug", cfg.LogLevel)
	assert.Equal(t, "debug", cfg.GinMode)
}

func TestWithEnvInvalidPort(t *testing.T) {
	_ = os.Setenv("PORT", "not-a-number")
	defer func() {
		_ = os.Unsetenv("PORT")
	}()

	_, err := NewBuilder().WithEnv().Validate().Build()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PORT must be a valid integer")
}

func TestValidatePortRange(t *testing.T) {
	tests := []struct {
		name    string
		port    string
		wantErr bool
		errMsg  string
	}{
		{"valid port", "8080", false, ""},
		{"port 1", "1", false, ""},
		{"port 65535", "65535", false, ""},
		{"port 0", "0", true, "PORT must be between 1 and 65535"},
		{"port negative", "-1", true, "PORT must be between 1 and 65535"},
		{"port too high", "65536", true, "PORT must be between 1 and 65535"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Setenv("PORT", tt.port)
			defer func() {
				_ = os.Unsetenv("PORT")
			}()

			_, err := NewBuilder().WithEnv().Validate().Build()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateLogLevel(t *testing.T) {
	tests := []struct {
		name     string
		logLevel string
		wantErr  bool
	}{
		{"debug", "debug", false},
		{"info", "info", false},
		{"warn", "warn", false},
		{"error", "error", false},
		{"invalid", "invalid", true},
		{"empty uses default", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Unsetenv("PORT")
			if tt.logLevel != "" {
				_ = os.Setenv("LOG_LEVEL", tt.logLevel)
			} else {
				_ = os.Unsetenv("LOG_LEVEL")
			}
			defer func() {
				_ = os.Unsetenv("LOG_LEVEL")
			}()

			_, err := NewBuilder().WithEnv().Validate().Build()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "LOG_LEVEL must be one of")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestValidateGinMode(t *testing.T) {
	tests := []struct {
		name    string
		ginMode string
		wantErr bool
	}{
		{"debug", "debug", false},
		{"release", "release", false},
		{"test", "test", false},
		{"invalid", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_ = os.Unsetenv("PORT")
			_ = os.Unsetenv("LOG_LEVEL")
			_ = os.Setenv("GIN_MODE", tt.ginMode)
			defer func() {
				_ = os.Unsetenv("GIN_MODE")
			}()

			_, err := NewBuilder().WithEnv().Validate().Build()

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "GIN_MODE must be one of")
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestMultipleErrors(t *testing.T) {
	_ = os.Setenv("PORT", "0")
	_ = os.Setenv("LOG_LEVEL", "invalid")
	_ = os.Setenv("GIN_MODE", "invalid")
	defer func() {
		_ = os.Unsetenv("PORT")
		_ = os.Unsetenv("LOG_LEVEL")
		_ = os.Unsetenv("GIN_MODE")
	}()

	_, err := NewBuilder().WithEnv().Validate().Build()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "PORT must be between 1 and 65535")
	assert.Contains(t, err.Error(), "LOG_LEVEL must be one of")
	assert.Contains(t, err.Error(), "GIN_MODE must be one of")
}

func TestBuildWithoutValidate(t *testing.T) {
	cfg, err := NewBuilder().WithEnv().Build()

	require.NoError(t, err)
	assert.NotNil(t, cfg)
}

func TestChainedCalls(t *testing.T) {
	_ = os.Unsetenv("PORT")
	_ = os.Unsetenv("LOG_LEVEL")
	_ = os.Unsetenv("GIN_MODE")

	builder := NewBuilder()
	builder = builder.WithEnv()
	builder = builder.Validate()
	cfg, err := builder.Build()

	require.NoError(t, err)
	assert.NotNil(t, cfg)
}
