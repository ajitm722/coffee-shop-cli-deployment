#!/usr/bin/env fish
# Simple interactive order creator for the Coffee API (fish shell)

set -l BASE "http://192.168.49.2:30090/v1"

read -P "Customer name: " customer
read -P "Items (comma separated): " items
# → example: latte, cappuccino, flat white

# Quote each item and join with commas -> "latte","cappuccino","flat white"
set -l quoted (string split , -- $items | string trim | sed 's/^/"/; s/$/"/' | paste -sd "," -)

# Build JSON payload
set -l payload (printf '{"customer":"%s","items":[%s]}' "$customer" "$quoted")

echo ""
echo "======================================"
echo "        PRINTING ORDER DOCKET"
echo "======================================"
echo $payload
echo "======================================"
echo ""

# POST to API
curl -s -X POST -H "Content-Type: application/json" \
  -d "$payload" \
  $BASE/orders | jq .

