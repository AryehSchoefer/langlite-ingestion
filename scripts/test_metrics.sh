#!/bin/bash

API_KEY="test-key-123"
BASE_URL="http://localhost:8080"

echo "📊 Testing Prometheus Metrics & Monitoring"
echo "=========================================="
echo ""

check_services() {
  echo "🔍 Checking Services:"

  if curl -s "$BASE_URL/health" >/dev/null; then
    echo "  ✅ API Server: Running"
  else
    echo "  ❌ API Server: Not responding"
    return 1
  fi

  if curl -s "http://localhost:9090/-/healthy" >/dev/null; then
    echo "  ✅ Prometheus: Running (http://localhost:9090)"
  else
    echo "  ❌ Prometheus: Not responding"
  fi

  if curl -s "http://localhost:3000/api/health" >/dev/null; then
    echo "  ✅ Grafana: Running (http://localhost:3000)"
  else
    echo "  ❌ Grafana: Not responding"
  fi

  echo ""
}

test_metrics_endpoint() {
  echo "📈 Testing Metrics Endpoint:"

  response=$(curl -s "$BASE_URL/metrics")
  if [ $? -eq 0 ]; then
    echo "  ✅ Metrics endpoint accessible"

    if echo "$response" | grep -q "http_requests_total"; then
      echo "  ✅ HTTP request metrics found"
    else
      echo "  ⚠️  HTTP request metrics not found"
    fi

    if echo "$response" | grep -q "queue_depth"; then
      echo "  ✅ Queue metrics found"
    else
      echo "  ⚠️  Queue metrics not found"
    fi

    if echo "$response" | grep -q "worker_status"; then
      echo "  ✅ Worker metrics found"
    else
      echo "  ⚠️  Worker metrics not found"
    fi

    metric_count=$(echo "$response" | grep -c "^[a-zA-Z]")
    echo "  📊 Total metrics: $metric_count"
  else
    echo "  ❌ Metrics endpoint not accessible"
  fi

  echo ""
}

generate_traffic() {
  echo "🚀 Generating Traffic for Metrics:"

  echo "  Making 10 API requests..."
  for i in {1..10}; do
    curl -s -X POST "$BASE_URL/api/v1/traces" \
      -H "Authorization: Bearer $API_KEY" \
      -H "Content-Type: application/json" \
      -d '{"name": "metrics-test-trace-'$i'"}' >/dev/null

    if [ $((i % 3)) -eq 0 ]; then
      echo -n "."
    fi
  done
  echo " Done!"

  echo "  ⏳ Waiting 5 seconds for metrics to update..."
  sleep 5
  echo ""
}

check_prometheus_targets() {
  echo "🎯 Checking Prometheus Targets:"

  targets=$(curl -s "http://localhost:9090/api/v1/targets" 2>/dev/null)
  if [ $? -eq 0 ]; then
    echo "  ✅ Prometheus API accessible"

    if echo "$targets" | grep -q "langlite-api"; then
      echo "  ✅ LangLite API target configured"

      if echo "$targets" | grep -q '"health":"up"'; then
        echo "  ✅ API target is healthy"
      else
        echo "  ⚠️  API target may be down"
      fi
    else
      echo "  ❌ LangLite API target not found"
    fi
  else
    echo "  ❌ Prometheus API not accessible"
  fi

  echo ""
}

show_sample_queries() {
  echo "📊 Sample Prometheus Queries:"
  echo ""
  echo "  HTTP Request Rate:"
  echo "    rate(http_requests_total[5m])"
  echo ""
  echo "  Queue Depth:"
  echo "    queue_depth"
  echo ""
  echo "  Worker Status:"
  echo "    worker_status"
  echo ""
  echo "  Request Latency (95th percentile):"
  echo "    histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))"
  echo ""
  echo "  Error Rate:"
  echo "    rate(http_requests_total{status_code=~\"5..\"}[5m])"
  echo ""
}

show_access_info() {
  echo "🌐 Access Information:"
  echo "  📊 Prometheus: http://localhost:9090"
  echo "  📈 Grafana: http://localhost:3000 (admin/admin)"
  echo "  🔧 Metrics Endpoint: http://localhost:8080/metrics"
  echo "  📋 API Health: http://localhost:8080/health"
  echo ""
  echo "💡 Next Steps:"
  echo "  1. Open Grafana and explore the dashboards"
  echo "  2. Create custom dashboards for your metrics"
  echo "  3. Set up alerting rules in Prometheus"
  echo "  4. Consider migrating to managed Prometheus when ready"
  echo ""
}

# Main execution
echo "1 Service Health Check"
check_services

echo "2 Metrics Endpoint Test"
test_metrics_endpoint

echo "3 Traffic Generation"
generate_traffic

echo "4 Prometheus Integration"
check_prometheus_targets

echo "5 Sample Queries"
show_sample_queries

echo "6 Access Information"
show_access_info

echo "✨ Metrics Testing Complete!"

