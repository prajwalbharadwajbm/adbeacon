package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	kitlog "github.com/go-kit/log"
	"github.com/prajwalbharadwajbm/adbeacon/internal/config"
	"github.com/prajwalbharadwajbm/adbeacon/internal/endpoint"
	"github.com/prajwalbharadwajbm/adbeacon/internal/middleware"
	"github.com/prajwalbharadwajbm/adbeacon/internal/repository"
	"github.com/prajwalbharadwajbm/adbeacon/internal/service"
	"github.com/prajwalbharadwajbm/adbeacon/internal/transport"
)

const VERSION = "1.0.0"

func init() {
	config.LoadConfigs()
	log.Println("AdBeacon: Loaded all configs")
}

func main() {
	// Create go-kit logger
	logger := kitlog.NewLogfmtLogger(os.Stderr)
	logger = kitlog.With(logger, "ts", kitlog.DefaultTimestampUTC)
	logger = kitlog.With(logger, "caller", kitlog.DefaultCaller)
	logger = kitlog.With(logger, "service", "adbeacon", "version", VERSION)

	// Repository layer (data access)
	campaignRepo := repository.NewMockRepository()

	// Service layer with middleware
	var deliveryService service.DeliveryService
	deliveryService = service.NewDeliveryService(campaignRepo)
	deliveryService = middleware.NewLoggingMiddleware(logger)(deliveryService)

	// Endpoint layer (request/response handling)
	endpoints := endpoint.MakeDeliveryEndpoints(deliveryService)

	// Transport layer (HTTP)
	httpHandler := transport.NewHTTPHandler(endpoints, logger)

	// HTTP server configuration
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.AppConfigInstance.GeneralConfig.Port),
		Handler:      httpHandler,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Server startup
	log.Printf("AdBeacon server starting on port %d", config.AppConfigInstance.GeneralConfig.Port)
	log.Println("Available endpoints:")
	log.Println("   GET /v1/delivery - Campaign delivery endpoint")
	log.Println("   GET /health      - Health check endpoint")

	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
