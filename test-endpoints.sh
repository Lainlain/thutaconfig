#!/bin/bash

# Test App Config Server Endpoints

BASE_URL="http://localhost:8080"

echo "🧪 Testing Burma 2D App Config Server"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Test 1: Health Check
echo ""
echo "1️⃣  Testing Health Check..."
curl -s "$BASE_URL/health" | jq

# Test 2: Main Config Endpoint
echo ""
echo "2️⃣  Testing Main Config Endpoint..."
curl -s "$BASE_URL/api/burma2d/config" | jq

# Test 3: Version Endpoint
echo ""
echo "3️⃣  Testing Version Endpoint..."
curl -s "$BASE_URL/api/burma2d/version" | jq

# Test 4: Messages Endpoint
echo ""
echo "4️⃣  Testing Messages Endpoint..."
curl -s "$BASE_URL/api/burma2d/messages" | jq

# Test 5: Root Endpoint
echo ""
echo "5️⃣  Testing Root Endpoint..."
curl -s "$BASE_URL/" | jq

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ All tests complete!"
