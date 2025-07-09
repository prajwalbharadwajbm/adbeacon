# Build stage
FROM golang:1.23-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o adbeacon ./cmd/server

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN addgroup -g 1000 adbeacon && adduser -u 1000 -G adbeacon -s /bin/sh -D adbeacon

# Set working directory
WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/adbeacon .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Change ownership to non-root user
RUN chown -R adbeacon:adbeacon /app

# Switch to non-root user
USER adbeacon

# Expose the application port
EXPOSE 8080

# Run the application
CMD ["./adbeacon"] 