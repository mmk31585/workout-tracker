package config

import (
	"fmt"
	"os"
	"strconv"
	"sync"

	"github.com/joho/godotenv"
)

var (
	cfg     *Config
	loadErr error
	once    sync.Once
)

type Config struct {
	App       AppConfig
	HttpPort  string
	HttpHost  string
	DB        DBConfig
	LogLevel  string
	LogFormat string
	JWT       JWTConfig
}
type AppConfig struct {
	AppEnv  string
	AppName string
}
type DBConfig struct {
	DBAddr       string
	MaxOpenConns int
	MaxIdleConns int
	MaxIdleTime  string
}
type JWTConfig struct {
	JWTSecret string
	JWTExp    string
}

func init() {
	err := godotenv.Load()
	if err != nil {
		loadErr = err
		return
	}
}

func Load() (*Config, error) {
	once.Do(func() {
		appEnv := GetString("APP_ENV", "development")
		appName := GetString("APP_NAME", "Workout Tracker")

		httpPort := GetString("HTTP_PORT", "8080")
		httpHost := GetString("HTTP_HOST", "0.0.0.0")

		dbAddr := GetString("DB_ADDR", "")
		maxOpenConns := GetInt("DB_MAX_OPEN_CONNS", 30)
		maxIdleConns := GetInt("DB_MAX_IDLE_CONNS", 30)
		maxIdleTime := GetString("DB_MAX_LIFE_TIME", "5m")

		logLevel := GetString("LOG_LEVEL", "info")
		logFormat := GetString("LOG_FORMAT", "")

		jwtSecret := GetString("JWT_SECRET", "change-me")
		jwtExp := GetString("JWT_EXPIRATION", "1h")

		if dbAddr == "" {
			loadErr = fmt.Errorf("config: DB_ADDR is required")
			return
		}

		if logFormat == "" {
			if appEnv == "production" {
				logFormat = "json"
			} else {
				logFormat = "text"
			}
		}
		if len(jwtSecret) < 32 {
			loadErr = fmt.Errorf("JWT_SECRET too short")
			return
		}
		dbConfig := &DBConfig{
			DBAddr:       dbAddr,
			MaxOpenConns: maxOpenConns,
			MaxIdleConns: maxIdleConns,
			MaxIdleTime:  maxIdleTime,
		}
		jwtConfig := &JWTConfig{
			JWTSecret: jwtSecret,
			JWTExp:    jwtExp,
		}
		appConfig := &AppConfig{
			AppEnv:  appEnv,
			AppName: appName,
		}
		cfg = &Config{
			App:       *appConfig,
			HttpPort:  httpPort,
			HttpHost:  httpHost,
			DB:        *dbConfig,
			LogLevel:  logLevel,
			LogFormat: logFormat,
			JWT:       *jwtConfig,
		}
	})

	return cfg, loadErr
}
func GetConfig() *Config {
	return cfg
}
func GetString(key string, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return val
}
func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	valAsInt, err := strconv.Atoi(val)

	if err != nil {
		return fallback
	}
	return valAsInt
}
