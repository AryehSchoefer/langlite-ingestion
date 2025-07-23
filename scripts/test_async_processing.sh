#!/bin/bash

API_KEY="test-key-123"
BASE_URL="http://localhost:8080"

echo "🚀 Testing Async Processing Pipeline"
echo "===================================="
echo ""

check_processing_status() {
  echo "📊 Processing Status:"
  curl -s -X GET "$BASE_URL/processing-status" -H "Authorization: Bearer $API_KEY" |
    python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    print(f'  Async Processing: {data.get(\"async_processing_enabled\", False)}')
    print(f'  Redis Connected: {data.get(\"redis_connected\", False)}')
    print(f'  Processing Mode: {data.get(\"processing_mode\", \"unknown\")}')
    print(f'  Response Status: {data.get(\"response_status\", \"unknown\")}')
except Exception as e:
    print(f'  Error: {e}')
"
  echo ""
}

check_queue_status() {
  echo "📋 Queue Status:"
  curl -s -X GET "$BASE_URL/queue-status" -H "Authorization: Bearer $API_KEY" |
    python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if data.get('queue_enabled'):
        queues = data.get('queues', {})
        summary = data.get('summary', {})
        print(f'  Total Pending: {summary.get(\"total_pending\", 0)}')
        print(f'  High Priority: {summary.get(\"high_priority\", 0)}')
        print(f'  Medium Priority: {summary.get(\"medium_priority\", 0)}')
        print(f'  Low Priority: {summary.get(\"low_priority\", 0)}')
        print(f'  Dead Letter: {queues.get(\"dead_letter\", 0)}')
        print(f'  Delayed: {queues.get(\"delayed\", 0)}')
    else:
        print('  Queue system disabled')
except Exception as e:
    print(f'  Error: {e}')
"
  echo ""
}

check_worker_status() {
  echo "👷 Worker Status:"
  curl -s -X GET "$BASE_URL/worker-status" -H "Authorization: Bearer $API_KEY" |
    python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if data.get('workers_enabled'):
        stats = data.get('stats', {})
        print(f'  Worker Count: {stats.get(\"worker_count\", 0)}')
        workers = stats.get('workers', [])
        running_count = sum(1 for w in workers if w.get('running', False))
        print(f'  Running Workers: {running_count}')
    else:
        print('  Workers disabled')
except Exception as e:
    print(f'  Error: {e}')
"
  echo ""
}

create_test_trace() {
  echo "🔄 Creating test trace..."
  response=$(curl -s -w "HTTP_CODE:%{http_code}" -X POST "$BASE_URL/api/v1/traces" \
    -H "Authorization: Bearer $API_KEY" \
    -H "Content-Type: application/json" \
    -d '{"name": "async-test-trace", "metadata": {"test": true}}')

  http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
  body=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')

  if [ "$http_code" = "202" ]; then
    echo "  ✅ Trace accepted for async processing (HTTP 202)"
    trace_id=$(echo "$body" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    print(data.get('id', ''))
except:
    pass
")
    echo "  📝 Trace ID: $trace_id"
  else
    echo "  ❌ Failed to create trace (HTTP $http_code)"
    echo "  Response: $body"
  fi
  echo ""
}

monitor_redis_queues() {
  echo "🔍 Redis Queue Contents:"

  if ! docker exec langlite-ingestion-redis-1 redis-cli ping >/dev/null 2>&1; then
    echo "  ❌ Redis not accessible"
    return
  fi

  keys=$(docker exec langlite-ingestion-redis-1 redis-cli keys "*:*" 2>/dev/null | grep -E "(high|medium|low|jobs)" | head -10)

  if [ -z "$keys" ]; then
    echo "  📭 No queue keys found"
  else
    echo "  📦 Found queue keys:"
    for key in $keys; do
      if [[ $key == *"jobs:tracking"* ]]; then
        echo "    🏷️  $key (tracking)"
      elif [[ $key == *"jobs:completed"* ]]; then
        echo "    ✅ $key (completed)"
      elif [[ $key == *"jobs:dead_letter"* ]]; then
        echo "    💀 $key (dead letter)"
      else
        length=$(docker exec langlite-ingestion-redis-1 redis-cli llen "$key" 2>/dev/null || echo "0")
        echo "    📋 $key: $length items"
      fi
    done
  fi
  echo ""
}

# Main test sequence
echo "1 Inital Status Check"
check_processing_status
check_queue_status
check_worker_status

echo "2 Creating Test Data"
create_test_trace

echo "3 Monitoring Processing"
echo "⏳ Waiting 2 seconds for processing..."
sleep 2

check_queue_status
monitor_redis_queues

echo "4 Final Status Check"
check_processing_status

echo "✨ Async Processing Test Complete!"
echo ""
echo "💡 Tips:"
echo "  - HTTP 202 responses indicate async processing"
echo "  - HTTP 201 responses indicate sync processing (fallback)"
echo "  - Monitor queue-status endpoint for job progress"
echo "  - Check worker-status for worker health"

