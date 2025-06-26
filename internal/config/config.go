package config

import (
	"log"

	"github.com/joho/godotenv"
	"github.com/prajwalbharadwajbm/adbeacon/internal/utils"
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
		log.Println("Error loading .env files: ", err)
	}

	loadGeneralCongigs()
}

var AppConfigInstance appConfig

// loadGeneralCongigs loads the general configurations from the environment variables
func loadGeneralCongigs() {
	AppConfigInstance.GeneralConfig.Env = utils.GetEnv("APP_ENV", "dev")
	AppConfigInstance.GeneralConfig.LogLevel = utils.GetEnv("LOG_LEVEL", "info")
	AppConfigInstance.GeneralConfig.Port = utils.GetEnv("PORT", 8080)
}
