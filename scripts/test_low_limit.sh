#!/bin/bash

echo "Testing with low rate limits to verify rate limiting works..."
echo "Making 10 requests quickly - should hit a low limit"

for i in {1..10}; do
  echo -n "Request $i: "

  response=$(curl -s -w "HTTP_CODE:%{http_code}" -X POST http://localhost:8080/api/v1/traces \
    -H "Authorization: Bearer test-key-123" \
    -H "Content-Type: application/json" \
    -d '{"name": "test-trace-'$i'"}')

  http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)

  if [ "$http_code" = "201" ]; then
    echo "SUCCESS"
  elif [ "$http_code" = "429" ]; then
    echo "RATE LIMITED (429)"
    body=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')
    echo "  Response: $body"
  else
    echo "ERROR ($http_code)"
    body=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')
    echo "  Response: $body"
  fi
done

echo ""
echo "Current Redis keys:"
docker exec langlite-ingestion-redis-1 redis-cli keys "*rate_limit*"

echo ""
echo "Rate limit values:"
for key in $(docker exec langlite-ingestion-redis-1 redis-cli keys "*rate_limit*"); do
  value=$(docker exec langlite-ingestion-redis-1 redis-cli get "$key")
  echo "  $key: $value"
done

