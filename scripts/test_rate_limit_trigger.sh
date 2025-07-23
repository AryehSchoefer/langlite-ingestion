#!/bin/bash

echo "Testing rate limiting by making 1100 requests (should trigger 1000/minute limit)..."

success_count=0
rate_limited_count=0
error_count=0

# Make requests in parallel to trigger rate limiting faster
for i in {1..1100}; do
  {
    response=$(curl -s -w "HTTP_CODE:%{http_code}" -X POST http://localhost:8080/api/v1/traces \
      -H "Authorization: Bearer test-key-123" \
      -H "Content-Type: application/json" \
      -d '{"name": "test-trace-'$i'"}')

    http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)

    if [ "$http_code" = "201" ]; then
      echo "Request $i: SUCCESS"
    elif [ "$http_code" = "429" ]; then
      echo "Request $i: RATE LIMITED (429)"
    else
      echo "Request $i: ERROR ($http_code)"
    fi
  } &

  # Limit concurrent requests to avoid overwhelming the system
  if ((i % 50 == 0)); then
    wait
  fi
done

wait # Wait for all background jobs to complete

echo ""
echo "Test completed! Checking final rate limit status..."

# Check final status
curl -X GET http://localhost:8080/rate-limit-status -H "Authorization: Bearer test-key-123" 2>/dev/null |
  python3 -c "
import sys, json
try:
    data = json.load(sys.stdin)
    print('Final Rate Limit Status:')
    print(f'  Minute usage: {data[\"usage\"][\"current_minute\"]}/1000')
    print(f'  Hour usage: {data[\"usage\"][\"current_hour\"]}/10000')
    print(f'  Minute remaining: {data[\"remaining\"][\"current_minute\"]}')
    print(f'  Hour remaining: {data[\"remaining\"][\"current_hour\"]}')
except:
    print('Could not parse rate limit status')
"

