package context

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RequestContextKey represents keys used in request context
type RequestContextKey string

const (
	// RequestIDKey is the context key for request ID
	RequestIDKey RequestContextKey = "request_id"
	// StartTimeKey is the context key for request start time
	StartTimeKey RequestContextKey = "start_time"
	// UserAgentKey is the context key for user agent
	UserAgentKey RequestContextKey = "user_agent"
	// RemoteAddrKey is the context key for remote address
	RemoteAddrKey RequestContextKey = "remote_addr"
)

// RequestInfo holds information about the current request
type RequestInfo struct {
	ID         string    `json:"request_id"`
	StartTime  time.Time `json:"start_time"`
	UserAgent  string    `json:"user_agent,omitempty"`
	RemoteAddr string    `json:"remote_addr,omitempty"`
}

// WithRequestID adds a request ID to the context
func WithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, RequestIDKey, requestID)
}

// GetRequestID retrieves the request ID from context
func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}

// WithStartTime adds a start time to the context
func WithStartTime(ctx context.Context, startTime time.Time) context.Context {
	return context.WithValue(ctx, StartTimeKey, startTime)
}

// GetStartTime retrieves the start time from context
func GetStartTime(ctx context.Context) time.Time {
	if startTime, ok := ctx.Value(StartTimeKey).(time.Time); ok {
		return startTime
	}
	return time.Time{}
}

// WithUserAgent adds user agent to the context
func WithUserAgent(ctx context.Context, userAgent string) context.Context {
	return context.WithValue(ctx, UserAgentKey, userAgent)
}

// GetUserAgent retrieves the user agent from context
func GetUserAgent(ctx context.Context) string {
	if userAgent, ok := ctx.Value(UserAgentKey).(string); ok {
		return userAgent
	}
	return ""
}

// WithRemoteAddr adds remote address to the context
func WithRemoteAddr(ctx context.Context, remoteAddr string) context.Context {
	return context.WithValue(ctx, RemoteAddrKey, remoteAddr)
}

// GetRemoteAddr retrieves the remote address from context
func GetRemoteAddr(ctx context.Context) string {
	if remoteAddr, ok := ctx.Value(RemoteAddrKey).(string); ok {
		return remoteAddr
	}
	return ""
}

// NewRequestContext creates a new request context with all necessary information
func NewRequestContext(ctx context.Context, userAgent, remoteAddr string) context.Context {
	requestID := uuid.New().String()
	startTime := time.Now()

	ctx = WithRequestID(ctx, requestID)
	ctx = WithStartTime(ctx, startTime)
	ctx = WithUserAgent(ctx, userAgent)
	ctx = WithRemoteAddr(ctx, remoteAddr)

	return ctx
}

// GetRequestInfo extracts all request information from context
func GetRequestInfo(ctx context.Context) RequestInfo {
	return RequestInfo{
		ID:         GetRequestID(ctx),
		StartTime:  GetStartTime(ctx),
		UserAgent:  GetUserAgent(ctx),
		RemoteAddr: GetRemoteAddr(ctx),
	}
}
