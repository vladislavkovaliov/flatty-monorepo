#!/bin/sh
set -e

BASE_URL="${SMOKE_URL:-http://localhost}"

echo "Smoke testing $BASE_URL"

check() {
  local url="$1"
  local label="$2"
  local status=$(curl -s -o /dev/null -w "%{http_code}" "$url" 2>/dev/null || echo "000")
  if [ "$status" = "000" ]; then
    echo "  FAIL: $label ($url) — no response"
    exit 1
  elif [ "$status" -ge 200 ] && [ "$status" -lt 500 ]; then
    echo "  OK:   $label ($status)"
  else
    echo "  FAIL: $label ($url) — HTTP $status"
    exit 1
  fi
}

echo ""
echo "=== Frontends ==="
check "$BASE_URL/" "Wrapper"
check "$BASE_URL/launcher/" "Launcher"

echo ""
echo "=== APIs ==="
check "$BASE_URL/api/health" "Go API health"
check "$BASE_URL/api/categories" "Go API categories"
check "$BASE_URL/api/resident-location" "Go API resident-location"
check "$BASE_URL/api/expenses" "Go API expenses"

echo ""
echo "=== GraphQL ==="
curl -s -X POST "$BASE_URL/graphql" \
  -H "Content-Type: application/json" \
  -d '{"query":"{ __typename }"}' > /dev/null && echo "  OK:   GraphQL" || (echo "  FAIL: GraphQL" && exit 1)

echo ""
echo "All smoke tests passed"
