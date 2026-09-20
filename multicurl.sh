#!/bin/bash

for i in $(seq -w 3001 8000); do
  curl -s -o /dev/null -w "item$i -> %{http_code}\n" \
    -X POST http://localhost:8005/api/v1/add-data \
    -H "Content-Type: application/json" \
    -d "{\"key\": \"item$i\", \"value\": \"item$i\"}" &
done
wait