#!/bin/bash

echo "Starting AdBeacon Monitoring Stack..."
echo ""

# Start the monitoring stack
echo "Starting Prometheus and Grafana..."
docker compose -f docker-compose.monitoring.yml up -d

echo ""
echo "Monitoring stack is starting up!"
echo ""
echo "Access URLs:"
echo "   Grafana Dashboard: http://localhost:3000"
echo "      Username: admin"
echo "      Password: admin123"
echo ""
echo "   Prometheus UI: http://localhost:9090"
echo "   AdBeacon App: http://localhost:8080"
echo "   AdBeacon Metrics: http://localhost:8080/metrics"
echo ""
echo "Wait ~30 seconds for everything to start, then open Grafana!"
echo ""
echo "To stop: docker-compose -f docker-compose.monitoring.yml down" 