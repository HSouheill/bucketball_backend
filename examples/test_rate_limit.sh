#!/bin/bash

# Rate Limiting Test Script for BucketBall Backend
# This script demonstrates the professional rate limiting system

BASE_URL="http://localhost:8080"
ADMIN_TOKEN="your-admin-token-here"  # Replace with actual admin token

echo "🚀 Testing BucketBall Backend Rate Limiting System"
echo "=================================================="

# Test 1: Normal login attempt (should fail with wrong password)
echo -e "\n📝 Test 1: Normal failed login attempt"
curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"wrongpassword"}' \
  | jq '.'

# Test 2: Multiple failed attempts to trigger rate limit
echo -e "\n📝 Test 2: Triggering rate limit (5 failed attempts)"
for i in {1..6}; do
  echo "Attempt $i:"
  curl -s -X POST "$BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com","password":"wrongpassword"}' \
    | jq '.message'
  sleep 1
done

# Test 3: Check rate limit status (requires admin token)
echo -e "\n📝 Test 3: Checking rate limit status (Admin only)"
if [ "$ADMIN_TOKEN" != "your-admin-token-here" ]; then
  curl -s -X GET "$BASE_URL/api/v1/admin/rate-limit/info?email=test@example.com" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    | jq '.'
else
  echo "⚠️  Admin token not set. Skipping admin tests."
fi

# Test 4: Reset rate limit (requires admin token)
echo -e "\n📝 Test 4: Resetting rate limit (Admin only)"
if [ "$ADMIN_TOKEN" != "your-admin-token-here" ]; then
  curl -s -X POST "$BASE_URL/api/v1/admin/rate-limit/reset?email=test@example.com" \
    -H "Authorization: Bearer $ADMIN_TOKEN" \
    | jq '.'
else
  echo "⚠️  Admin token not set. Skipping reset test."
fi

# Test 5: Test with different IP (simulate different client)
echo -e "\n📝 Test 5: Testing with different IP simulation"
curl -s -X POST "$BASE_URL/api/v1/auth/login" \
  -H "Content-Type: application/json" \
  -H "X-Forwarded-For: 192.168.1.100" \
  -d '{"email":"test@example.com","password":"wrongpassword"}' \
  | jq '.'

echo -e "\n✅ Rate limiting test completed!"
echo "💡 Expected behavior:"
echo "   - First 5 attempts: 401 Unauthorized"
echo "   - 6th attempt: 429 Too Many Requests"
echo "   - Admin endpoints: Rate limit info and reset"
