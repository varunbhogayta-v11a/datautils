#!/bin/bash
set -e

echo "=== DataUtil Testing Setup ==="

cd "$(dirname "$0")"

echo ""
echo "Step 1: Starting PostgreSQL..."
if command -v pg_ctl &> /dev/null; then
    pg_ctl -D /var/lib/postgresql/data start 2>/dev/null || \
    sudo -u postgres pg_ctl -D /var/lib/postgresql/data start 2>/dev/null || \
    echo "Please start PostgreSQL manually: sudo systemctl start postgresql"
fi

echo ""
echo "Step 2: Creating database config..."
cat > .env << 'EOF'
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=datautil_test
JWT_SECRET=test-secret-key-for-testing
EOF

echo ""
echo "Step 3: Running database initialization..."
export $(cat .env | xargs)
./datautil init-db

echo ""
echo "Step 4: Creating test user..."
./datautil register --username testuser --email test@example.com --password test123

echo ""
echo "Step 5: Getting JWT token..."
TOKEN=$(./datautil login --email test@example.com --password test123 2>/dev/null | grep -oE '[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+\.[a-zA-Z0-9_-]+' | head -1)

if [ -z "$TOKEN" ]; then
    echo "Error: Could not get token"
    exit 1
fi

echo "Token obtained: ${TOKEN:0:50}..."

echo ""
echo "=== Testing Commands ==="

echo ""
echo "Test 1: Filter with auth..."
./datautil filter --input tests/test_data.csv --where "age > 25" --token "$TOKEN"

echo ""
echo "Test 2: Transform with auth..."
./datautil transform --input tests/test_data.csv --add "test_col=name" --token "$TOKEN"

echo ""
echo "Test 3: Validate with auth..."
./datautil validate --input tests/test_data.csv --required name,email --token "$TOKEN"

echo ""
echo "Test 4: Export with auth..."
./datautil export --input tests/test_data.csv --to json --output /tmp/test_export.json --token "$TOKEN"

echo ""
echo "Test 5: View logs..."
./datautil logs --token "$TOKEN"

echo ""
echo "Test 6: List users (admin)..."
./datautil users --token "$TOKEN"

echo ""
echo "=== All tests passed! ==="
echo ""
echo "Quick reference:"
echo "  Token: $TOKEN"
echo "  Save this token to use in future commands"
echo ""
echo "Example usage:"
echo "  ./datautil filter --input data.csv --where 'age > 25' --token '$TOKEN'"