#!/bin/bash

# Start date
start_date=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# File to write the entries to
file="entries.json"

# Clear the file
> $file

# Generate 3000 entries
for i in $(seq 1 3000); do
  # Calculate the end date by adding i minutes to the start date
  # The -v flag is used to add i minutes to the start date. Use -d for GNU date (Linux) or -v for BSD date (macOS)
  end_date=$(date -u -v+${i}M +"%Y-%m-%dT%H:%M:%SZ")

  # Generate the entry
  entry=$(cat <<EOF
{
  "title": "Ad $i",
  "startAt": "$start_date",
  "endAt": "$end_date",
  "conditions": {
            "ageStart": 1,
            "ageEnd": 100
    }
}
EOF
)

  # Write the entry to the file
  echo $entry >> $file
done

# Read the target URL, method, body, and headers from config.json
TARGET_URL=$(jq -r '.url' config.json)
METHOD=$(jq -r '.method' config.json)

# Read the headers from config.json into an array
HEADERS=()
while IFS= read -r line; do
  HEADERS+=("$line")
done < <(jq -r '.headers | to_entries[] | "\(.key): \(.value)"' config.json)

# File to read the test data from
file="entries.json"

# Read the test data from the file and send it to the target URL
while IFS= read -r BODY; do
  curl -X "${METHOD}" "${TARGET_URL}" \
    "${HEADERS[@]/#/-H}" \
    -d "$BODY" > /dev/null 2>&1
done < $file