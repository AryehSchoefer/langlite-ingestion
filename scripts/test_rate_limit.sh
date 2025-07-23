#!/bin/bash

echo "Testing rate limiting..."
echo "Making 10 requests quickly to test rate limiting"

for i in {1..10}; do
  echo "Request $i:"
  response=$(curl -s -w "HTTP_CODE:%{http_code}" -X POST http://localhost:8080/api/v1/traces \
    -H "Authorization: Bearer test-key-123" \
    -H "Content-Type: application/json" \
    -d '{"name": "test-trace-'$i'"}')

  http_code=$(echo "$response" | grep -o "HTTP_CODE:[0-9]*" | cut -d: -f2)
  body=$(echo "$response" | sed 's/HTTP_CODE:[0-9]*$//')

  echo "  Status: $http_code"
  echo "  Response: $body"
  echo ""

  # Small delay to see rate limiting in action
  sleep 0.1
done

echo "Testing completed!"

