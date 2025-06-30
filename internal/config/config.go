package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type GeneralConfig struct {
	Env      string
	LogLevel string
	Port     int
}

type DatabaseConfig struct {
	Host            string
	Port            int
	User            string
	Password        string
	DBName          string
	SSLMode         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int // in minutes
	ConnMaxIdleTime int // in minutes
}

type appConfig struct {
	GeneralConfig  GeneralConfig
	DatabaseConfig DatabaseConfig
}

// LoadConfigs loads the configurations from the environment variables
func LoadConfigs() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env files: %v", err)
	}

	loadGeneralConfigs()
	loadDatabaseConfigs()
}

var AppConfigInstance appConfig

// loadGeneralConfigs loads the general configurations from the environment variables
func loadGeneralConfigs() {
	AppConfigInstance.GeneralConfig.Env = getEnv("APP_ENV", "dev")
	AppConfigInstance.GeneralConfig.LogLevel = getEnv("LOG_LEVEL", "info")
	AppConfigInstance.GeneralConfig.Port = getEnvInt("PORT", 8080)
}

// loadDatabaseConfigs loads the database configurations from the environment variables
func loadDatabaseConfigs() {
	AppConfigInstance.DatabaseConfig.Host = getEnv("DB_HOST", "localhost")
	AppConfigInstance.DatabaseConfig.Port = getEnvInt("DB_PORT", 5432)
	AppConfigInstance.DatabaseConfig.User = getEnv("DB_USER", "adbeacon_dev_user")
	AppConfigInstance.DatabaseConfig.Password = getEnv("DB_PASSWORD", "")
	AppConfigInstance.DatabaseConfig.DBName = getEnv("DB_NAME", "adbeacon")
	AppConfigInstance.DatabaseConfig.SSLMode = getEnv("DB_SSLMODE", "disable")
	AppConfigInstance.DatabaseConfig.MaxOpenConns = getEnvInt("DB_MAX_OPEN_CONNS", 25)
	AppConfigInstance.DatabaseConfig.MaxIdleConns = getEnvInt("DB_MAX_IDLE_CONNS", 25)
	AppConfigInstance.DatabaseConfig.ConnMaxLifetime = getEnvInt("DB_CONN_MAX_LIFETIME", 5)
	AppConfigInstance.DatabaseConfig.ConnMaxIdleTime = getEnvInt("DB_CONN_MAX_IDLE_TIME", 5)
}

// getEnv returns the environment variable value if it exists, otherwise returns the fallback value
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvInt returns the environment variable value as int if it exists, otherwise returns the fallback value
func getEnvInt(key string, fallback int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return fallback
}
