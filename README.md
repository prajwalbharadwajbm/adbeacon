# AdBeacon - Ad Delivery System

A high-performance ad delivery system with hybrid caching and comprehensive monitoring.

## Prerequisites

- Go 1.21+
- Docker and Docker Compose
- curl or Postman

## Setup

1. Clone the repository
```bash
git clone <repository-url>
cd adbeacon
```

2. Start infrastructure services
```bash
docker compose -f docker-compose.yml -f docker-compose.monitoring.yml up -d
```

3. Verify services are running
```bash
docker ps
```

4. Start the application
```bash
go run ./cmd/server
```

The server will start on port 8080.

## API Endpoints

### Campaign Delivery
```
GET /v1/delivery?country={country}&os={os}&app={app}
```
- `country`: 2-letter country code (required)
- `os`: Operating system - android/ios (required)  
- `app`: Application package name (required)

### Health Check
```
GET /health
```

### Metrics
```
GET /metrics
```

## Testing

### Valid Requests - (data is added to the database and cache on startup)

**Get Spotify campaign (US users):**
```bash
curl "http://localhost:8080/v1/delivery?country=US&os=Android&app=com.example.testapp"
```

**Get Duolingo campaign (non-US users):**
```bash
curl "http://localhost:8080/v1/delivery?country=CA&os=iOS&app=com.example.testapp"
```

**Get Subway Surfer campaign (specific app):**
```bash
curl "http://localhost:8080/v1/delivery?country=IN&os=Android&app=com.gametion.ludokinggame"
```

**Get multiple campaigns:**
```bash
curl "http://localhost:8080/v1/delivery?country=CA&os=Android&app=com.gametion.ludokinggame"
```

### Validation Errors

**Missing parameters:**
```bash
curl "http://localhost:8080/v1/delivery"
# Returns: 400 {"error":"country is required"}
```

**Invalid country code:**
```bash
curl "http://localhost:8080/v1/delivery?country=INVALID&os=Android&app=test"
# Returns: 400 {"error":"country must be a 2-letter code"}
```

### Health and Metrics

**Check system health:**
```bash
curl "http://localhost:8080/health"
```

**View metrics:**
```bash
curl "http://localhost:8080/metrics"
```

## Expected Response Format

**Successful delivery:**
```json
[
  {
    "cid": "spotify",
    "img": "https://somelink",
    "cta": "Download"
  }
]
```

**No matching campaigns:**
```
HTTP 204 No Content
```

**Validation error:**
```json
{
  "error": "country is required"
}
```

## Performance

- **First request:** ~50-100ms (cache warming)
- **Cached requests:** <1ms
- **Cache hit ratio:** 90%+
- **Database queries:** Minimal (2 queries for all requests)

## Campaign Targeting Rules

Current campaigns and their targeting:

| Campaign | Country | OS | App |
|----------|---------|----|----|
| Spotify | US, Canada | Any | Any |
| Duolingo | Non-US | Android, iOS | Any |
| Subway Surfer | Any | Android | com.gametion.ludokinggame |

## Infrastructure Services

- **PostgreSQL:** localhost:5432
- **Redis:** localhost:6379  
- **Adminer:** localhost:8081
- **Grafana:** localhost:3000

## Stopping Services

```bash
# Stop application
Ctrl+C

# Stop infrastructure
docker compose -f docker-compose.yml -f docker-compose.monitoring.yml down
```