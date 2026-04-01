#!/bin/bash

# AppConfig Native Ad Timer - Test Script
# Tests the dynamic ad timer configuration endpoints

echo "🧪 Testing AppConfig Native Ad Timer Endpoints"
echo "================================================"
echo ""

# Check if server is running
echo "1️⃣ Checking server health..."
curl -s http://localhost:8080/health | jq .status
echo ""

# Get current ad config
echo "2️⃣ Getting current ad config..."
CURRENT=$(curl -s http://localhost:8080/api/burma2d/adconfig)
echo "$CURRENT" | jq .
echo ""

# Update to 90 seconds
echo "3️⃣ Updating ad timer to 90 seconds..."
curl -s -X POST http://localhost:8080/api/burma2d/adconfig \
  -H "Content-Type: application/json" \
  -d '{"native_ad_timer_seconds": 90}' | jq .
echo ""

# Verify update
echo "4️⃣ Verifying update..."
curl -s http://localhost:8080/api/burma2d/adconfig | jq .
echo ""

# Test full config includes ad_config
echo "5️⃣ Checking full config includes ad_config..."
curl -s http://localhost:8080/api/burma2d/config | jq '.ad_config'
echo ""

# Update to 120 seconds
echo "6️⃣ Updating ad timer to 120 seconds..."
curl -s -X POST http://localhost:8080/api/burma2d/adconfig \
  -H "Content-Type: application/json" \
  -d '{"native_ad_timer_seconds": 120}' | jq .
echo ""

# Verify
echo "7️⃣ Verifying..."
curl -s http://localhost:8080/api/burma2d/adconfig | jq .
echo ""

# Reset to default 60 seconds
echo "8️⃣ Resetting to default (60 seconds)..."
curl -s -X POST http://localhost:8080/api/burma2d/adconfig \
  -H "Content-Type: application/json" \
  -d '{"native_ad_timer_seconds": 60}' | jq .
echo ""

# Final check
echo "9️⃣ Final verification..."
curl -s http://localhost:8080/api/burma2d/adconfig | jq .
echo ""

echo "✅ All tests completed!"
echo ""
echo "📊 Summary:"
echo "  - GET  /api/burma2d/adconfig  → Get ad timer"
echo "  - POST /api/burma2d/adconfig  → Update ad timer"
echo "  - GET  /api/burma2d/config    → Full config (includes ad_config)"
