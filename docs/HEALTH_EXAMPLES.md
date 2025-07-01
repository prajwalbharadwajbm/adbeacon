# Health Endpoint Examples

The `/health` endpoint now includes comprehensive cache health information along with database health.

## Healthy System (All Components Working)

```bash
curl http://localhost:8080/health
```

```json
{
  "status": "healthy",
  "service": "adbeacon", 
  "version": "1.0.0",
  "database": {
    "status": "healthy",
    "stats": {
      "open_connections": 5,
      "in_use": 1,
      "idle": 4,
      "wait_count": 0,
      "wait_duration": "0s",
      "max_idle_closed": 0,
      "max_idle_time_closed": 0,
      "max_lifetime_closed": 0
    }
  },
  "cache": {
    "overall": "healthy",
    "memory": {
      "enabled": true,
      "status": "healthy", 
      "size": 15,
      "max_size": 1000,
      "util_pct": 1.5,
      "evicted_keys": 0
    },
    "redis": {
      "enabled": true,
      "status": "healthy",
      "connected": true,
      "address": "localhost:6379",
      "latency": "2.1ms"
    },
    "stats": {
      "hits": 15420,
      "misses": 250,
      "errors": 0,
      "hit_ratio": 0.984,
      "total_ops": 15670,
      "last_updated": "2024-01-15T10:30:45Z"
    },
    "uptime": "45m30s",
    "last_test": "2024-01-15T10:35:12Z"
  }
}
```

**HTTP Status**: `200 OK`

---

## Degraded System (High Cache Utilization)

```json
{
  "status": "degraded",
  "service": "adbeacon",
  "version": "1.0.0", 
  "database": {
    "status": "healthy",
    "stats": { /* ... */ }
  },
  "cache": {
    "overall": "degraded",
    "memory": {
      "enabled": true,
      "status": "degraded",
      "size": 950,
      "max_size": 1000,
      "util_pct": 95.0,
      "evicted_keys": 15
    },
    "redis": {
      "enabled": true,
      "status": "degraded",
      "connected": true,
      "address": "localhost:6379",
      "latency": "75ms"
    },
    "stats": {
      "hits": 89230,
      "misses": 4520,
      "errors": 12,
      "hit_ratio": 0.952,
      "total_ops": 93750
    },
    "uptime": "2h15m",
    "last_test": "2024-01-15T10:35:12Z"
  }
}
```

**HTTP Status**: `200 OK` (degraded but functional)

---

## Unhealthy System (Redis Down)

```json
{
  "status": "unhealthy", 
  "service": "adbeacon",
  "version": "1.0.0",
  "database": {
    "status": "healthy",
    "stats": { /* ... */ }
  },
  "cache": {
    "overall": "unhealthy",
    "memory": {
      "enabled": true,
      "status": "healthy",
      "size": 125,
      "max_size": 1000,
      "util_pct": 12.5,
      "evicted_keys": 0
    },
    "redis": {
      "enabled": true,
      "status": "unhealthy",
      "connected": false,
      "address": "localhost:6379",
      "latency": "0s",
      "error": "dial tcp 127.0.0.1:6379: connect: connection refused"
    },
    "stats": {
      "hits": 5240,
      "misses": 8950,
      "errors": 145,
      "hit_ratio": 0.369,
      "total_ops": 14190
    },
    "uptime": "12m45s",
    "last_test": "2024-01-15T10:35:12Z"
  }
}
```

**HTTP Status**: `503 Service Unavailable`

---

## Memory-Only Configuration

```json
{
  "status": "healthy",
  "service": "adbeacon",
  "version": "1.0.0",
  "database": {
    "status": "healthy",
    "stats": { /* ... */ }
  },
  "cache": {
    "overall": "healthy",
    "memory": {
      "enabled": true,
      "status": "healthy",
      "size": 42,
      "max_size": 1000,
      "util_pct": 4.2,
      "evicted_keys": 0
    },
    "redis": {
      "enabled": false,
      "status": "disabled",
      "connected": false,
      "address": "localhost:6379"
    },
    "stats": {
      "hits": 12450,
      "misses": 180,
      "errors": 0,
      "hit_ratio": 0.986,
      "total_ops": 12630
    },
    "uptime": "1h23m",
    "last_test": "2024-01-15T10:35:12Z"
  }
}
```

---

## Health Status Meanings

### Overall Status
- **`healthy`**: All enabled components working optimally
- **`degraded`**: Components working but with performance issues
- **`unhealthy`**: Critical components failing

### Memory Cache Status
- **`healthy`**: Low utilization (<90%), no evictions
- **`degraded`**: High utilization (>90%), frequent evictions  
- **`unhealthy`**: Cache failures or critical errors
- **`disabled`**: Memory caching turned off

### Redis Cache Status
- **`healthy`**: Connected, low latency (<50ms)
- **`degraded`**: Connected, high latency (>50ms)
- **`unhealthy`**: Connection failures or errors
- **`disabled`**: Redis caching turned off

---

## Alerting Thresholds

### Critical Alerts (Status = unhealthy)
- Database connection failed
- Redis connection failed (if enabled)
- Cache hit ratio < 70%
- Cache errors > 100/minute

### Warning Alerts (Status = degraded)  
- Memory cache utilization > 90%
- Redis latency > 50ms
- Cache hit ratio < 95%
- Cache errors > 10/minute

### Performance Monitoring
- **Hit Ratio**: Should be >95% for optimal performance
- **Memory Utilization**: Should be <90% to avoid evictions
- **Redis Latency**: Should be <10ms for best performance
- **Total Operations**: Tracks cache usage volume

---

## Using Health Data for Operations

### Load Balancer Health Checks
```bash
# Simple health check
curl -f http://adbeacon:8080/health

# Detailed health check with jq
curl -s http://adbeacon:8080/health | jq '.status'
```

### Prometheus Monitoring
```yaml
# prometheus.yml
- job_name: 'adbeacon-health'
  static_configs:
    - targets: ['adbeacon:8080']
  metrics_path: '/health'
  scrape_interval: 30s
```

### Kubernetes Readiness Probe
```yaml
readinessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 10
  periodSeconds: 30
```

### Docker Health Check
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1
``` 