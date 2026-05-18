#!/bin/bash

# datautils Master Demo Script
# This script demonstrates EVERY feature of the datautils tool.

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
PURPLE='\033[0;35m'
CYAN='\033[0;36m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}===========================================${NC}"
echo -e "${BLUE}       datautils Comprehensive Demo        ${NC}"
echo -e "${BLUE}===========================================${NC}"

# 0. Build the application
echo -e "\n${PURPLE}[0/12] Building datautil binary...${NC}"
go build -o datautil main.go
if [ $? -ne 0 ]; then
    echo -e "${RED}Build failed!${NC}"
    exit 1
fi
echo -e "Build successful."

# 1. Clean up and Initialize Database
echo -e "\n${GREEN}[1/12] Initializing Database...${NC}"
rm -f datautil.db
./datautil init-db

# 2. User Authentication
echo -e "\n${GREEN}[2/12] Registering and Promoting Admin...${NC}"
./datautil register -u admin -e admin@example.com -p admin123
# Promote to admin role manually since register defaults to 'user'
sqlite3 datautil.db "UPDATE users SET role='admin' WHERE username='admin';"
TOKEN=$(./datautil login -e admin@example.com -p admin123 | grep -A 1 "JWT Token:" | tail -n 1 | xargs)

if [ -z "$TOKEN" ]; then
    echo -e "${RED}Failed to capture JWT Token${NC}"
    exit 1
fi
echo -e "Authenticated as Admin (Role: admin)."

# 3. Data Import
echo -e "\n${GREEN}[3/12] Importing Titanic Dataset...${NC}"
./datautil import --input tests/titanic.csv --table passengers --create --token "$TOKEN"

# 4. Data Processing (Filter & Transform)
echo -e "\n${GREEN}[4/12] Filtering and Transforming Data...${NC}"
echo -e "${CYAN}Filtering (Age > 40) and selecting columns...${NC}"
./datautil filter --input tests/titanic.csv --where "Age > 40" --select Name,Age,Sex,Fare --output processed.csv --token "$TOKEN"
echo -e "${CYAN}Renaming Sex to Gender...${NC}"
./datautil transform --input processed.csv --rename "Sex:Gender" --output processed.csv --token "$TOKEN"
echo -e "Top 3 processed rows:"
head -n 4 processed.csv

# 5. Data Validation
echo -e "\n${GREEN}[5/12] Validating Data Schema...${NC}"
./datautil validate --input processed.csv --required Name,Age,Gender --token "$TOKEN"

# 6. Data Export
echo -e "\n${GREEN}[6/12] Exporting to JSON and XML...${NC}"
./datautil export --input processed.csv --to json --output results.json --token "$TOKEN"
./datautil export --input processed.csv --to xml --output results.xml --token "$TOKEN"
echo -e "Exported files: results.json, results.xml"

# 7. SQL Querying
echo -e "\n${GREEN}[7/12] Executing SQL Queries...${NC}"
./datautil query --sql "SELECT COUNT(*) as Count, Gender FROM (SELECT 'male' as Gender UNION SELECT 'female') GROUP BY Gender" --token "$TOKEN"

# 8. CRUD Operations
echo -e "\n${GREEN}[8/12] Performing CRUD Operations (Insert/Update/Delete)...${NC}"
echo -e "${CYAN}Inserting new record...${NC}"
./datautil insert --table passengers --values "Name=Demo User,Age=25,Sex=male,Fare=50.0" --token "$TOKEN"
echo -e "${CYAN}Updating record...${NC}"
./datautil update --table passengers --set "Age=26" --where "Name='Demo User'" --token "$TOKEN"
echo -e "${CYAN}Verifying update...${NC}"
./datautil query --sql "SELECT Name, Age FROM passengers WHERE Name='Demo User'" --token "$TOKEN"
echo -e "${CYAN}Deleting record...${NC}"
./datautil delete --table passengers --where "Name='Demo User'" --token "$TOKEN"

# 9. System Logs
echo -e "\n${GREEN}[9/12] Viewing Operation Logs...${NC}"
./datautil logs --token "$TOKEN" | tail -n 5

# 10. User Management
echo -e "\n${GREEN}[10/12] Listing Registered Users...${NC}"
./datautil users --token "$TOKEN"

# 11. API Server Demo
echo -e "\n${GREEN}[11/12] API Server Demonstration...${NC}"
echo -e "${CYAN}Starting server in background on port 8080...${NC}"
./datautil server --port 8080 > server.log 2>&1 &
SERVER_PID=$!
sleep 2 # Wait for server to start

echo -e "Server started (PID: $SERVER_PID). Testing API with curl..."
echo -e "${CYAN}API: Health Check...${NC}"
curl -s http://localhost:8080/api/health
echo -e "\n${CYAN}API: List Users...${NC}"
curl -s -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/users | head -c 100 && echo "..."

echo -e "\n${CYAN}Stopping server...${NC}"
kill $SERVER_PID
wait $SERVER_PID 2>/dev/null

# 12. Cleanup
echo -e "\n${GREEN}[12/12] Cleaning up...${NC}"
rm -f datautil processed.csv results.json results.xml server.log
echo -e "${YELLOW}Demo finished! Database 'datautil.db' remains for your inspection.${NC}"

echo -e "\n${BLUE}===========================================${NC}"
echo -e "${BLUE}       Demo Completed Successfully         ${NC}"
echo -e "${BLUE}===========================================${NC}"
