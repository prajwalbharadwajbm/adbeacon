package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/prajwalbharadwajbm/adbeacon/internal/config"
	"github.com/rs/zerolog/log"
)

const VERSION = "1.0.0"

func init() {
	config.LoadConfigs()
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

	log.Info().Msgf("Starting server on port %d", config.AppConfigInstance.GeneralConfig.Port)
	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal().Err(err).Msg("failed to serve http server")
	}
}
