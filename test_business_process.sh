#!/bin/bash

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="http://localhost:8080/api/v1"
API_KEY="be-parkir-api-key-2025"
HEADERS="-H 'X-API-Key: $API_KEY' -H 'Content-Type: application/json'"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Testing Be-Parkir Business Process${NC}"
echo -e "${GREEN}========================================${NC}\n"

# Step 1: Register Admin User
echo -e "${YELLOW}Step 1: Register Admin User${NC}"
ADMIN_REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Admin User",
    "email": "admin@parkir.com",
    "phone": "081234567890",
    "password": "admin123456",
    "role": "admin"
  }')

echo "$ADMIN_REGISTER_RESPONSE" | jq . 2>/dev/null || echo "$ADMIN_REGISTER_RESPONSE"
ADMIN_TOKEN=$(echo "$ADMIN_REGISTER_RESPONSE" | jq -r '.data.access_token // empty' 2>/dev/null)
echo -e "${GREEN}Admin Token: ${ADMIN_TOKEN:0:50}...${NC}\n"

# Step 2: Register Jukir User
echo -e "${YELLOW}Step 2: Register Jukir User${NC}"
JUKIR_REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Edi Mangcek",
    "email": "edi.jukir@parkir.com",
    "phone": "081234567891",
    "password": "jukir123456",
    "role": "jukir"
  }')

echo "Response: $JUKIR_REGISTER_RESPONSE" | jq .
JUKIR_USER_ID=$(echo $JUKIR_REGISTER_RESPONSE | jq -r '.data.user.id')
JUKIR_TOKEN=$(echo $JUKIR_REGISTER_RESPONSE | jq -r '.data.access_token')
echo -e "${GREEN}Jukir User ID: $JUKIR_USER_ID${NC}"
echo -e "${GREEN}Jukir Token: ${JUKIR_TOKEN:0:50}...${NC}\n"

# Step 3: Register Customer User
echo -e "${YELLOW}Step 3: Register Customer User${NC}"
CUSTOMER_REGISTER_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/register" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Customer User",
    "email": "customer@parkir.com",
    "phone": "081234567892",
    "password": "customer123456",
    "role": "customer"
  }')

echo "Response: $CUSTOMER_REGISTER_RESPONSE" | jq .
CUSTOMER_TOKEN=$(echo $CUSTOMER_REGISTER_RESPONSE | jq -r '.data.access_token')
echo -e "${GREEN}Customer Token: ${CUSTOMER_TOKEN:0:50}...${NC}\n"

# Step 4: Admin creates parking area
echo -e "${YELLOW}Step 4: Admin creates parking area${NC}"
AREA_RESPONSE=$(curl -s -X POST "$BASE_URL/admin/areas" \
  -H "X-API-Key: $API_KEY" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Kambang Iwak Selatan",
    "address": "Jl. Kambang Iwak Selatan, Palembang",
    "latitude": -2.9834,
    "longitude": 104.7604,
    "hourly_rate": 5000
  }')

echo "Response: $AREA_RESPONSE" | jq .
AREA_ID=$(echo $AREA_RESPONSE | jq -r '.data.id')
echo -e "${GREEN}Area ID: $AREA_ID${NC}\n"

# Step 5: Admin creates jukir profile
echo -e "${YELLOW}Step 5: Admin creates jukir profile${NC}"
JUKIR_PROFILE_RESPONSE=$(curl -s -X POST "$BASE_URL/admin/jukirs" \
  -H "X-API-Key: $API_KEY" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"user_id\": $JUKIR_USER_ID,
    \"jukir_code\": \"JK21\",
    \"area_id\": $AREA_ID
  }")

echo "Response: $JUKIR_PROFILE_RESPONSE" | jq .
JUKIR_ID=$(echo $JUKIR_PROFILE_RESPONSE | jq -r '.data.id')
echo -e "${GREEN}Jukir Profile ID: $JUKIR_ID${NC}\n"

# Step 6: Admin updates jukir status to active
echo -e "${YELLOW}Step 6: Admin activates jukir${NC}"
UPDATE_JUKIR_RESPONSE=$(curl -s -X PUT "$BASE_URL/admin/jukirs/$JUKIR_ID/status" \
  -H "X-API-Key: $API_KEY" \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "status": "active"
  }')

echo "Response: $UPDATE_JUKIR_RESPONSE" | jq .

# Step 7: Jukir needs to login to get new token
echo -e "${YELLOW}Step 7: Jukir login with role to get new token${NC}"
JUKIR_LOGIN_RESPONSE=$(curl -s -X POST "$BASE_URL/auth/login" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "email": "edi.jukir@parkir.com",
    "password": "jukir123456"
  }')

echo "Response: $JUKIR_LOGIN_RESPONSE" | jq .
JUKIR_ACCESS_TOKEN=$(echo $JUKIR_LOGIN_RESPONSE | jq -r '.data.access_token')
echo -e "${GREEN}Jukir Access Token: ${JUKIR_ACCESS_TOKEN:0:50}...${NC}\n"

# Step 8: Get Jukir QR Code
echo -e "${YELLOW}Step 8: Get Jukir QR Code${NC}"
QR_RESPONSE=$(curl -s -X GET "$BASE_URL/jukir/qr-code" \
  -H "X-API-Key: $API_KEY" \
  -H "Authorization: Bearer $JUKIR_ACCESS_TOKEN")

echo "Response: $QR_RESPONSE" | jq .
QR_TOKEN=$(echo $QR_RESPONSE | jq -r '.data.qr_token')
echo -e "${GREEN}QR Token: $QR_TOKEN${NC}\n"

# Step 9: Customer check-in
echo -e "${YELLOW}Step 9: Customer check-in${NC}"
CHECKIN_RESPONSE=$(curl -s -X POST "$BASE_URL/parking/checkin" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"qr_token\": \"$QR_TOKEN\",
    \"vehicle_type\": \"motor\",
    \"plat_nomor\": \"B1234XYZ\"
  }")

echo "Response: $CHECKIN_RESPONSE" | jq .
SESSION_ID=$(echo $CHECKIN_RESPONSE | jq -r '.data.session_id')
echo -e "${GREEN}Session ID: $SESSION_ID${NC}\n"

# Step 10: Customer checkout
echo -e "${YELLOW}Step 10: Customer checkout${NC}"
# Wait a bit before checkout
sleep 2

CHECKOUT_RESPONSE=$(curl -s -X POST "$BASE_URL/parking/checkout" \
  -H "X-API-Key: $API_KEY" \
  -H "Content-Type: application/json" \
  -d "{
    \"session_id\": $SESSION_ID
  }")

echo "Response: $CHECKOUT_RESPONSE" | jq .

# Step 11: Get admin overview (includes last 7 days revenue)
echo -e "${YELLOW}Step 11: Get admin overview with 7 days revenue data${NC}"
OVERVIEW_RESPONSE=$(curl -s -X GET "$BASE_URL/admin/overview" \
  -H "X-API-Key: $API_KEY" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

echo "Response:" | jq .
echo "$OVERVIEW_RESPONSE" | jq '.data | {today_revenue, vehicles_in, vehicles_out, last_7days_revenue, jukir_status}'

# Step 12: Get jukirs with revenue
echo -e "${YELLOW}Step 12: Get jukirs with revenue${NC}"
JUKIRS_REVENUE_RESPONSE=$(curl -s -X GET "$BASE_URL/admin/jukirs?include_revenue=true" \
  -H "X-API-Key: $API_KEY" \
  -H "Authorization: Bearer $ADMIN_TOKEN")

echo "Response:" | jq .
echo "$JUKIRS_REVENUE_RESPONSE" | jq '.data[] | {jukir_code, user: .user.name, revenue, status}'

# Step 13: Get jukir dashboard
echo -e "${YELLOW}Step 13: Get jukir dashboard${NC}"
JUKIR_DASHBOARD_RESPONSE=$(curl -s -X GET "$BASE_URL/jukir/dashboard" \
  -H "X-API-Key: $API_KEY" \
  -H "Authorization: Bearer $JUKIR_ACCESS_TOKEN")

echo "Response: $JUKIR_DASHBOARD_RESPONSE" | jq .

echo -e "\n${GREEN}========================================${NC}"
echo -e "${GREEN}Test Completed Successfully!${NC}"
echo -e "${GREEN}========================================${NC}"

