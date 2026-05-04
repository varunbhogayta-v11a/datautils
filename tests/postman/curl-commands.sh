#!/bin/bash
# DataUtil API - Quick Test Commands
# Usage: Source this file or copy commands

BASE_URL="http://localhost:8080"

echo "=== DataUtil API Quick Test Commands ==="
echo ""

echo "1. Health Check:"
echo "curl $BASE_URL/api/health"
echo ""

echo "2. Register User:"
echo "curl -X POST $BASE_URL/api/auth/register \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"username\":\"testuser\",\"email\":\"test@example.com\",\"password\":\"test123\"}'"
echo ""

echo "3. Login:"
echo "curl -X POST $BASE_URL/api/auth/login \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -d '{\"email\":\"test@example.com\",\"password\":\"test123\"}'"
echo ""

echo "4. Filter Data (replace TOKEN with actual token):"
echo "curl -X POST $BASE_URL/api/data/filter \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -H 'Authorization: Bearer TOKEN' \\"
echo "  -d '{\"input\":\"tests/test_data.csv\",\"where\":\"age > 25\"}'"
echo ""

echo "5. Transform Data:"
echo "curl -X POST $BASE_URL/api/data/transform \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -H 'Authorization: Bearer TOKEN' \\"
echo "  -d '{\"input\":\"tests/test_data.csv\",\"add\":\"fullname=first+last\"}'"
echo ""

echo "6. Validate Data:"
echo "curl -X POST $BASE_URL/api/data/validate \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -H 'Authorization: Bearer TOKEN' \\"
echo "  -d '{\"input\":\"tests/test_data.csv\",\"required\":\"name,age\"}'"
echo ""

echo "7. Export Data:"
echo "curl -X POST $BASE_URL/api/data/export \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -H 'Authorization: Bearer TOKEN' \\"
echo "  -d '{\"input\":\"tests/test_data.csv\",\"to\":\"json\",\"output\":\"output.json\"}'"
echo ""

echo "8. DB Query:"
echo "curl -X POST $BASE_URL/api/db/query \\"
echo "  -H 'Content-Type: application/json' \\"
echo "  -H 'Authorization: Bearer TOKEN' \\"
echo "  -d '{\"sql\":\"SELECT * FROM users LIMIT 5\"}'"
echo ""

echo "9. List Users:"
echo "curl $BASE_URL/api/users \\"
echo "  -H 'Authorization: Bearer TOKEN'"
echo ""

echo "=== Automated Quick Test ==="

# Test Health
echo -e "\n[TEST] Health Check..."
curl -s $BASE_URL/api/health | jq .

# Register test user
echo -e "\n[TEST] Register User..."
REGISTER_RESP=$(curl -s -X POST $BASE_URL/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"username":"apitest","email":"apitest@example.com","password":"test123"}')
echo $REGISTER_RESP | jq .

# Login and get token
echo -e "\n[TEST] Login..."
LOGIN_RESP=$(curl -s -X POST $BASE_URL/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"apitest@example.com","password":"test123"}')
echo $LOGIN_RESP | jq .

# Extract token (requires jq)
TOKEN=$(echo $LOGIN_RESP | jq -r '.data.token' 2>/dev/null)
if [ "$TOKEN" != "null" ] && [ -n "$TOKEN" ]; then
  echo -e "\n[SUCCESS] Got token: ${TOKEN:0:20}..."
  
  # Test Filter
  echo -e "\n[TEST] Filter Data..."
  curl -s -X POST $BASE_URL/api/data/filter \
    -H 'Content-Type: application/json' \
    -H "Authorization: Bearer $TOKEN" \
    -d '{"input":"tests/test_data.csv","where":"age > 25"}' | jq .
  
  # Test Query
  echo -e "\n[TEST] DB Query..."
  curl -s -X POST $BASE_URL/api/db/query \
    -H 'Content-Type: application/json' \
    -H "Authorization: Bearer $TOKEN" \
    -d '{"sql":"SELECT id, username, email FROM users LIMIT 3"}' | jq .
  
  # Test Users
  echo -e "\n[TEST] List Users..."
  curl -s $BASE_URL/api/users \
    -H "Authorization: Bearer $TOKEN" | jq .
else
  echo -e "\n[INFO] Login may have failed - using guest token for tests"
fi

echo -e "\n=== Quick Test Complete ==="