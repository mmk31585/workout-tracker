package config

import (
	"os"
	"strings"
	"testing"
)

const (
	validSecret = "0123456789abcdef0123456789abcdef"
	shortSecret = "0123456789abcdef0123456789abcde"
	testDBAddr  = "postgres://localhost:5432/test?sslmode=disable"
)

func clearEnv(t *testing.T) {
	t.Helper()

	old := os.Environ()
	os.Clearenv()
	t.Cleanup(func() {
		for _, kv := range old {
			if i := strings.IndexByte(kv, '='); i > 0 {
				_ = os.Setenv(kv[:i], kv[i+1:])
			}
		}
	})
}

func setBaseEnv(t *testing.T) {
	t.Helper()

	clearEnv(t)
	os.Setenv("DB_ADDR", testDBAddr)
	os.Setenv("JWT_SECRET", validSecret)
}

func TestLoad_AllDefaults(t *testing.T) {
	setBaseEnv(t)

	ResetForTest()
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error with defaults, got: %v", err)
	}

	if cfg.App.AppEnv != "development" {
		t.Errorf("App.AppEnv: expected development, got %q", cfg.App.AppEnv)
	}
	if cfg.App.AppName != "Workout Tracker" {
		t.Errorf("App.AppName: expected Workout Tracker, got %q", cfg.App.AppName)
	}
	if cfg.Http.HttpPort != "8080" {
		t.Errorf("Http.HttpPort: expected 8080, got %q", cfg.Http.HttpPort)
	}
	if cfg.Http.HttpHost != "0.0.0.0" {
		t.Errorf("Http.HttpHost: expected 0.0.0.0, got %q", cfg.Http.HttpHost)
	}
	if cfg.DB.DBAddr != testDBAddr {
		t.Errorf("DB.DBAddr: expected %q, got %q", testDBAddr, cfg.DB.DBAddr)
	}
	if cfg.DB.MaxOpenConns != 30 {
		t.Errorf("DB.MaxOpenConns: expected 30, got %d", cfg.DB.MaxOpenConns)
	}
	if cfg.DB.MaxIdleConns != 30 {
		t.Errorf("DB.MaxIdleConns: expected 30, got %d", cfg.DB.MaxIdleConns)
	}
	if cfg.DB.MaxIdleTime != "5m" {
		t.Errorf("DB.MaxIdleTime: expected 5m, got %q", cfg.DB.MaxIdleTime)
	}
	if cfg.Log.LogLevel != "info" {
		t.Errorf("Log.LogLevel: expected info, got %q", cfg.Log.LogLevel)
	}
	if cfg.Log.LogFormat != "text" {
		t.Errorf("Log.LogFormat: expected text in non-production, got %q", cfg.Log.LogFormat)
	}
	if cfg.JWT.JWTExp != "1h" {
		t.Errorf("JWT.JWTExp: expected 1h, got %q", cfg.JWT.JWTExp)
	}
}

func TestLoad_ProductionSetsJsonLogFormat(t *testing.T) {
	setBaseEnv(t)
	os.Setenv("APP_ENV", "production")

	ResetForTest()
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.Log.LogFormat != "json" {
		t.Errorf("expected LogFormat=json in production, got %q", cfg.Log.LogFormat)
	}
}

func TestLoad_CustomValues(t *testing.T) {
	clearEnv(t)
	os.Setenv("APP_ENV", "staging")
	os.Setenv("APP_NAME", "Custom Tracker")
	os.Setenv("HTTP_PORT", "9090")
	os.Setenv("HTTP_HOST", "127.0.0.1")
	os.Setenv("DB_ADDR", "postgres://user:pass@db:5432/mydb?sslmode=disable")
	os.Setenv("DB_MAX_OPEN_CONNS", "50")
	os.Setenv("DB_MAX_IDLE_CONNS", "10")
	os.Setenv("DB_MAX_LIFE_TIME", "10m")
	os.Setenv("LOG_LEVEL", "debug")
	os.Setenv("LOG_FORMAT", "json")
	os.Setenv("JWT_SECRET", validSecret)
	os.Setenv("JWT_EXPIRATION", "2h")

	ResetForTest()
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if cfg.App.AppEnv != "staging" {
		t.Errorf("App.AppEnv: expected staging, got %q", cfg.App.AppEnv)
	}
	if cfg.App.AppName != "Custom Tracker" {
		t.Errorf("App.AppName: expected Custom Tracker, got %q", cfg.App.AppName)
	}
	if cfg.Http.HttpPort != "9090" {
		t.Errorf("Http.HttpPort: expected 9090, got %q", cfg.Http.HttpPort)
	}
	if cfg.Http.HttpHost != "127.0.0.1" {
		t.Errorf("Http.HttpHost: expected 127.0.0.1, got %q", cfg.Http.HttpHost)
	}
	if cfg.DB.DBAddr != "postgres://user:pass@db:5432/mydb?sslmode=disable" {
		t.Errorf("DB.DBAddr: unexpected value %q", cfg.DB.DBAddr)
	}
	if cfg.DB.MaxOpenConns != 50 {
		t.Errorf("DB.MaxOpenConns: expected 50, got %d", cfg.DB.MaxOpenConns)
	}
	if cfg.DB.MaxIdleConns != 10 {
		t.Errorf("DB.MaxIdleConns: expected 10, got %d", cfg.DB.MaxIdleConns)
	}
	if cfg.DB.MaxIdleTime != "10m" {
		t.Errorf("DB.MaxIdleTime: expected 10m, got %q", cfg.DB.MaxIdleTime)
	}
	if cfg.Log.LogLevel != "debug" {
		t.Errorf("Log.LogLevel: expected debug, got %q", cfg.Log.LogLevel)
	}
	if cfg.Log.LogFormat != "json" {
		t.Errorf("Log.LogFormat: expected json, got %q", cfg.Log.LogFormat)
	}
	if cfg.JWT.JWTExp != "2h" {
		t.Errorf("JWT.JWTExp: expected 2h, got %q", cfg.JWT.JWTExp)
	}
}

func TestLoad_MissingDBAddrReturnsError(t *testing.T) {
	clearEnv(t)

	ResetForTest()
	cfg, err := Load()
	if err == nil {
		t.Fatal("expected error when DB_ADDR is not set, got nil")
	}
	if err.Error() != "config: DB_ADDR is required" {
		t.Errorf("expected specific error message, got: %v", err)
	}
	if cfg != nil {
		t.Errorf("expected nil config on error, got %+v", cfg)
	}
}

func TestLoad_EmptyDBAddrReturnsError(t *testing.T) {
	clearEnv(t)
	os.Setenv("DB_ADDR", "")
	os.Setenv("JWT_SECRET", validSecret)

	ResetForTest()
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when DB_ADDR is empty, got nil")
	}
}

func TestLoad_ShortJWTSecretReturnsError(t *testing.T) {
	setBaseEnv(t)
	os.Setenv("JWT_SECRET", shortSecret)

	ResetForTest()
	_, err := Load()
	if err == nil {
		t.Fatal("expected error when JWT_SECRET is shorter than 32 chars, got nil")
	}
	if !strings.Contains(err.Error(), "too short") {
		t.Errorf("expected 'too short' in error, got: %v", err)
	}
}

func TestLoad_JWTSecretExactly32CharsIsAccepted(t *testing.T) {
	setBaseEnv(t)

	ResetForTest()
	_, err := Load()
	if err != nil {
		t.Fatalf("expected exactly-32-char secret to be accepted, got: %v", err)
	}
}

func TestLoad_LogFormatMatrix(t *testing.T) {
	tests := []struct {
		name      string
		appEnv    string
		logFormat string
		want      string
	}{
		{"production auto defaults to json", "production", "", "json"},
		{"development auto defaults to text", "development", "", "text"},
		{"staging auto defaults to text", "staging", "", "text"},
		{"explicit format overrides production auto", "production", "text", "text"},
		{"explicit format wins in development", "development", "json", "json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			setBaseEnv(t)
			os.Setenv("APP_ENV", tt.appEnv)
			if tt.logFormat != "" {
				os.Setenv("LOG_FORMAT", tt.logFormat)
			}

			ResetForTest()
			cfg, err := Load()
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}

			if cfg.Log.LogFormat != tt.want {
				t.Errorf("expected LogFormat=%s, got %q", tt.want, cfg.Log.LogFormat)
			}
		})
	}
}

func TestLoad_InvalidIntFallsBackToDefault(t *testing.T) {
	setBaseEnv(t)
	os.Setenv("DB_MAX_OPEN_CONNS", "not-a-number")

	ResetForTest()
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected non-numeric int to fall back to default, got: %v", err)
	}

	if cfg.DB.MaxOpenConns != 30 {
		t.Errorf("expected MaxOpenConns=30 (default), got %d", cfg.DB.MaxOpenConns)
	}
}

func TestGetString_UnsetReturnsFallback(t *testing.T) {
	clearEnv(t)

	val := GetString("WT_TEST_NONEXISTENT_KEY", "fallback_value")
	if val != "fallback_value" {
		t.Errorf("expected fallback_value for unset key, got %q", val)
	}
}

func TestGetString_SetReturnsActualValue(t *testing.T) {
	t.Setenv("WT_TEST_STRING_KEY", "actual_value")

	val := GetString("WT_TEST_STRING_KEY", "fallback")
	if val != "actual_value" {
		t.Errorf("expected actual_value for set key, got %q", val)
	}
}

func TestGetString_EmptyValueIsReturnedAsSet(t *testing.T) {
	t.Setenv("WT_TEST_EMPTY_KEY", "")

	val := GetString("WT_TEST_EMPTY_KEY", "fallback")
	if val != "" {
		t.Errorf("expected empty string (LookupEnv semantics), got %q", val)
	}
}

func TestGetInt_UnsetReturnsFallback(t *testing.T) {
	clearEnv(t)

	val := GetInt("WT_TEST_NONEXISTENT_INT_KEY", 42)
	if val != 42 {
		t.Errorf("expected 42 for unset int key, got %d", val)
	}
}

func TestGetInt_SetReturnsParsedValue(t *testing.T) {
	t.Setenv("WT_TEST_INT_KEY", "123")

	val := GetInt("WT_TEST_INT_KEY", 0)
	if val != 123 {
		t.Errorf("expected 123 for set int key, got %d", val)
	}
}

func TestGetInt_NonNumericReturnsFallback(t *testing.T) {
	t.Setenv("WT_TEST_INT_KEY", "not-a-number")

	val := GetInt("WT_TEST_INT_KEY", 7)
	if val != 7 {
		t.Errorf("expected 7 (fallback) for non-numeric int key, got %d", val)
	}
}
