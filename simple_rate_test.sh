#!/bin/bash

echo "Simple rate limit test - making requests in a tight loop until rate limited..."

count=0
while true; do
  count=$((count + 1))

  response=$(curl -s -w "HTTP_CODE:%{http_code}" -X POST http://localhost:8080/api/v1/traces \
    -H "Authorization: Bearer test-key-123" \
    -H "Content-Type: application/json" \
    -d '{"name": "test-trace-'$count'"}')

  http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)

  if [ "$http_code" = "201" ]; then
    echo "Request $count: SUCCESS"
  elif [ "$http_code" = "429" ]; then
    echo "Request $count: RATE LIMITED! (429)"
    body=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')
    echo "Response: $body"
    echo ""
    echo "Rate limiting triggered after $count requests!"
    break
  else
    echo "Request $count: ERROR ($http_code)"
    body=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')
    echo "Response: $body"
    break
  fi

  # Stop after 1200 requests to avoid infinite loop
  if [ $count -ge 1200 ]; then
    echo "Reached 1200 requests without hitting rate limit"
    break
  fi
done

echo ""
echo "Final status check:"
curl -X GET http://localhost:8080/rate-limit-status -H "Authorization: Bearer test-key-123"

