package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/prajwalbharadwajbm/adbeacon/internal/handlers"
)

func Routes() *httprouter.Router {
	router := httprouter.New()
	router.HandlerFunc(http.MethodGet, "/health", handlers.Health)

	return router
}
