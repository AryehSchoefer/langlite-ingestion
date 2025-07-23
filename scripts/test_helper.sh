#!/bin/bash

API_KEY="test-key-123"
BASE_URL="http://localhost:8080"

show_help() {
  echo "Rate Limit Test Helper"
  echo ""
  echo "Usage: $0 [command]"
  echo ""
  echo "Commands:"
  echo "  status    - Show current rate limit status"
  echo "  reset     - Reset rate limits for current API key"
  echo "  test      - Run a simple test (5 requests)"
  echo "  stress    - Run stress test (50 requests)"
  echo "  trigger   - Test with low limits to trigger rate limiting"
  echo "  clear     - Clear all Redis data (nuclear option)"
  echo "  help      - Show this help"
  echo ""
}

check_status() {
  echo "Current rate limit status:"
  curl -s -X GET "$BASE_URL/rate-limit-status" -H "Authorization: Bearer $API_KEY" |
    python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    if data.get('rate_limiting_enabled'):
        print(f'  Minute: {data[\"usage\"][\"current_minute\"]}/1000 (remaining: {data[\"remaining\"][\"current_minute\"]})')
        print(f'  Hour: {data[\"usage\"][\"current_hour\"]}/10000 (remaining: {data[\"remaining\"][\"current_hour\"]})')
    else:
        print('  Rate limiting is disabled')
except Exception as e:
    print(f'  Error: {e}')
"
}

reset_limits() {
  echo "Resetting rate limits..."
  response=$(curl -s -X POST "$BASE_URL/reset-rate-limit" -H "Authorization: Bearer $API_KEY")
  echo "$response" | python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    print(f'  {data[\"message\"]}')
    print(f'  Keys deleted: {data[\"keys_deleted\"]}')
except:
    print('  Reset failed or not in development mode')
"
}

simple_test() {
  echo "Running simple test (5 requests)..."
  for i in {1..5}; do
    response=$(curl -s -w "HTTP_CODE:%{http_code}" -X POST "$BASE_URL/api/v1/traces" \
      -H "Authorization: Bearer $API_KEY" \
      -H "Content-Type: application/json" \
      -d '{"name": "test-trace-'$i'"}')

    http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
    echo "  Request $i: $([ "$http_code" = "201" ] && echo "SUCCESS" || echo "FAILED ($http_code)")"
  done
}

stress_test() {
  echo "Running stress test (50 requests)..."
  success=0
  failed=0

  for i in {1..50}; do
    response=$(curl -s -w "HTTP_CODE:%{http_code}" -X POST "$BASE_URL/api/v1/traces" \
      -H "Authorization: Bearer $API_KEY" \
      -H "Content-Type: application/json" \
      -d '{"name": "test-trace-'$i'"}')

    http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
    if [ "$http_code" = "201" ]; then
      success=$((success + 1))
    else
      failed=$((failed + 1))
    fi

    # Show progress every 10 requests
    if ((i % 10 == 0)); then
      echo "  Progress: $i/50 (Success: $success, Failed: $failed)"
    fi
  done

  echo "  Final: Success: $success, Failed: $failed"
}

clear_redis() {
  echo "Clearing all Redis data..."
  docker exec langlite-ingestion-redis-1 redis-cli flushall
  echo "  Redis cleared"
}

case "$1" in
"status")
  check_status
  ;;
"reset")
  reset_limits
  echo ""
  check_status
  ;;
"test")
  simple_test
  echo ""
  check_status
  ;;
"stress")
  stress_test
  echo ""
  check_status
  ;;
"trigger")
  echo "This would require temporarily modifying rate limits in code"
  echo "Use the test_low_limit.sh script instead"
  ;;
"clear")
  clear_redis
  ;;
"help" | "")
  show_help
  ;;
*)
  echo "Unknown command: $1"
  show_help
  exit 1
  ;;
esac

