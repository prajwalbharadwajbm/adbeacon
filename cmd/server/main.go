package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prajwalbharadwajbm/adbeacon/internal/config"
	"github.com/prajwalbharadwajbm/adbeacon/internal/logger"
)

const VERSION = "1.0.0"

func init() {
	config.LoadConfigs()
	initializeGlobalLogger()
	logger.Log.Info("loaded all configs")
}

func initializeGlobalLogger() {
	env := config.AppConfigInstance.GeneralConfig.Env
	logLevel := config.AppConfigInstance.GeneralConfig.LogLevel
	logger.InitializeGlobalLogger(logLevel, env, VERSION+"-adbeacon")
	logger.Log.Info("loaded the global logger")
}

func main() {
	router := Routes()

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.AppConfigInstance.GeneralConfig.Port),
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	logger.Log.Infof("Starting server on port %d", config.AppConfigInstance.GeneralConfig.Port)
	err := srv.ListenAndServe()
	if err != nil {
		logger.Log.Fatal("failed to serve http server", err)
	}
}
