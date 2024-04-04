#!/bin/bash

# Read the target URL, method, body, and headers from config.json
TARGET_URL=$(jq -r '.url' config.json)
METHOD=$(jq -r '.method' config.json)

# Read the headers from config.json into an array
HEADERS=()
while IFS= read -r line; do
  HEADERS+=("$line")
done < <(jq -r '.headers | to_entries[] | "\(.key): \(.value)"' config.json)

# Run the load test with vegeta
echo "$METHOD $TARGET_URL" | \
vegeta attack "${HEADERS[@]/#/-header=}" -max-connections=100 -rate=10000 -duration=10s | \
vegeta report -output="search_ad_report.txt"