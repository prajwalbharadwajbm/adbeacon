package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	kitlog "github.com/go-kit/log"
	"github.com/prajwalbharadwajbm/adbeacon/internal/cache"
	"github.com/prajwalbharadwajbm/adbeacon/internal/config"
	"github.com/prajwalbharadwajbm/adbeacon/internal/database"
	"github.com/prajwalbharadwajbm/adbeacon/internal/endpoint"
	"github.com/prajwalbharadwajbm/adbeacon/internal/metrics"
	"github.com/prajwalbharadwajbm/adbeacon/internal/middleware"
	"github.com/prajwalbharadwajbm/adbeacon/internal/repository"
	"github.com/prajwalbharadwajbm/adbeacon/internal/service"
	"github.com/prajwalbharadwajbm/adbeacon/internal/transport"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	// Initialize Prometheus metrics
	prometheusMetrics := metrics.NewPrometheusMetrics()
	log.Println("Prometheus metrics initialized")

	// Initialize database
	db, dbCleanup, err := database.Initialize(config.AppConfigInstance.DatabaseConfig, "./migrations")
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		log.Println("Closing database connection...")
		dbCleanup()
		log.Println("Database connection closed")
	}()
	log.Println("Database initialized successfully")

	// Add cache initialization example
	cache, err := initializeCache()
	if err != nil {
		log.Fatalf("Failed to initialize cache: %v", err)
	}
	log.Println("Cache initialized successfully")

	// Repository layer (data access) with caching
	cachedRepo := setupCachedRepository(db, cache)

	// Service layer with middleware
	var deliveryService service.CampaignDeliveryService
	deliveryService = service.NewDeliveryService(cachedRepo)
	deliveryService = middleware.NewServiceMetricsMiddleware(prometheusMetrics)(deliveryService)
	deliveryService = middleware.NewLoggingMiddleware(logger)(deliveryService)

	// Endpoint layer (request/response handling)
	endpoints := endpoint.MakeDeliveryEndpoints(deliveryService)

	// Transport layer (HTTP) with database and cache health checks
	httpHandler := transport.NewHTTPHandlerWithCache(endpoints, logger, db, cache)

	// Add request ID middleware (first in chain to ensure all requests have IDs)
	requestIDMiddleware := middleware.NewRequestIDMiddleware()
	httpHandler = requestIDMiddleware.Middleware(httpHandler)

	// Add metrics middleware to HTTP handler
	metricsMiddleware := middleware.NewMetricsMiddleware(prometheusMetrics)
	httpHandler = metricsMiddleware.Middleware(httpHandler)

	// Add Prometheus metrics endpoint
	http.Handle("/metrics", promhttp.Handler())
	http.Handle("/", httpHandler)

	// HTTP server configuration
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", config.AppConfigInstance.GeneralConfig.Port),
		Handler:      nil, // Using default ServeMux
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server in a goroutine so that it doesn't block the main thread
	go func() {
		log.Printf("AdBeacon server starting on port %d", config.AppConfigInstance.GeneralConfig.Port)
		log.Println("Available endpoints:")
		log.Println("   GET /v1/delivery - Campaign delivery endpoint")
		log.Println("   GET /health      - Health check endpoint")
		log.Println("   GET /metrics     - Prometheus metrics endpoint")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	} else {
		log.Println("Server exited gracefully")
	}
}

// Add cache initialization example
func initializeCache() (*cache.HybridCache, error) {
	cacheConfig := config.GetCacheConfig()

	hybridCache, err := cache.NewHybridCache(cacheConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize cache: %w", err)
	}

	return hybridCache, nil
}

// Add this to show how to wire up cached repository
func setupCachedRepository(db *database.DB, hybridCache *cache.HybridCache) service.CampaignRepository {
	// Original repository
	baseRepo := repository.NewPostgresRepository(db)

	// Wrap with instrumentation
	prometheusMetrics := metrics.NewPrometheusMetrics()
	instrumentedRepo := repository.NewInstrumentedRepository(baseRepo, prometheusMetrics)

	// Wrap with caching (5-minute TTL)
	cachedRepo := cache.NewCachedRepository(instrumentedRepo, hybridCache, 5*time.Minute)

	return cachedRepo
}
