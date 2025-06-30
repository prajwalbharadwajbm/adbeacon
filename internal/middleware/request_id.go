package middleware

import (
	"net/http"

	reqcontext "github.com/prajwalbharadwajbm/adbeacon/internal/context"
)

// RequestIDMiddleware adds request IDs to incoming requests
type RequestIDMiddleware struct{}

// NewRequestIDMiddleware creates a new request ID middleware
func NewRequestIDMiddleware() *RequestIDMiddleware {
	return &RequestIDMiddleware{}
}

// Middleware returns the HTTP middleware function for request IDs
func (m *RequestIDMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if request already has an ID from upstream (X-Request-ID header)
		existingRequestID := r.Header.Get("X-Request-ID")

		// Create request context with ID and metadata
		var ctx = r.Context()
		if existingRequestID != "" {
			ctx = reqcontext.WithRequestID(ctx, existingRequestID)
		} else {
			// Create new request context with generated ID
			ctx = reqcontext.NewRequestContext(ctx, r.UserAgent(), r.RemoteAddr)
		}

		// Get the request ID for response header
		requestID := reqcontext.GetRequestID(ctx)

		// Add request ID to response headers for client tracking
		w.Header().Set("X-Request-ID", requestID)

		// Continue with updated context
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
