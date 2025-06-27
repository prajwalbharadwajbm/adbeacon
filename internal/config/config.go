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

type appConfig struct {
	GeneralConfig GeneralConfig
}

// LoadConfigs loads the configurations from the environment variables
func LoadConfigs() {
	err := godotenv.Load()
	if err != nil {
		log.Printf("Warning: Error loading .env files: %v", err)
	}

	loadGeneralConfigs()
}

var AppConfigInstance appConfig

// loadGeneralConfigs loads the general configurations from the environment variables
func loadGeneralConfigs() {
	AppConfigInstance.GeneralConfig.Env = getEnv("APP_ENV", "dev")
	AppConfigInstance.GeneralConfig.LogLevel = getEnv("LOG_LEVEL", "info")
	AppConfigInstance.GeneralConfig.Port = getEnvInt("PORT", 8080)
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
