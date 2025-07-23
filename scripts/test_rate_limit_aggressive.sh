#!/bin/bash

echo "Testing rate limiting with aggressive requests..."
echo "Making 50 requests as fast as possible to trigger rate limiting"

success_count=0
rate_limited_count=0

for i in {1..50}; do
  response=$(curl -s -w "HTTP_CODE:%{http_code}" -X POST http://localhost:8080/api/v1/traces \
    -H "Authorization: Bearer test-key-123" \
    -H "Content-Type: application/json" \
    -d '{"name": "test-trace-'$i'"}')

  http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)

  if [ "$http_code" = "201" ]; then
    success_count=$((success_count + 1))
    echo "Request $i: SUCCESS"
  elif [ "$http_code" = "429" ]; then
    rate_limited_count=$((rate_limited_count + 1))
    echo "Request $i: RATE LIMITED (429)"
    body=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')
    echo "  Response: $body"
  else
    echo "Request $i: ERROR ($http_code)"
    body=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')
    echo "  Response: $body"
  fi
done

echo ""
echo "Results:"
echo "  Successful requests: $success_count"
echo "  Rate limited requests: $rate_limited_count"
echo "  Total requests: 50"

# Check Redis keys
echo ""
echo "Current Redis rate limit keys:"
docker exec langlite-ingestion-redis-1 redis-cli keys "*rate_limit*"

echo ""
echo "Rate limit values:"
for key in $(docker exec langlite-ingestion-redis-1 redis-cli keys "*rate_limit*"); do
  value=$(docker exec langlite-ingestion-redis-1 redis-cli get "$key")
  echo "  $key: $value"
done

