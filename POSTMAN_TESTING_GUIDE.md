# Be-Parkir API - Complete Postman Testing Guide

## 🚀 **Production Server Information**

- **Base URL**: `http://103.208.137.57/alulz`
- **API Version**: v1
- **Full API Base**: `http://103.208.137.57/alulz/api/v1`
- **Swagger UI**: `http://103.208.137.57/alulz/swagger/index.html`
- **Health Check**: `http://103.208.137.57/alulz/health`

## 🔑 **Authentication Requirements**

### **API Key Authentication**

All API requests require an API key in the header:

```
X-API-Key: be-parkir-api-key-2025
```

### **JWT Token Authentication**

For protected endpoints, include the JWT token:

```
Authorization: Bearer <your_jwt_token>
```

## 📋 **Complete Testing Workflow**

### **Phase 1: Authentication Flow**

#### **1.1 User Registration**

```http
POST http://103.208.137.57/alulz/api/v1/auth/register
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025

{
  "name": "John Doe",
  "email": "john.doe@example.com",
  "password": "password123",
  "phone": "081234567890",
  "role": "customer"
}
```

**Expected Response:**

```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "user": {
      "id": 1,
      "name": "John Doe",
      "email": "john.doe@example.com",
      "phone": "081234567890",
      "role": "customer",
      "status": "active",
      "created_at": "2025-10-25T07:35:26.253156536Z",
      "updated_at": "2025-10-25T07:35:26.253156536Z"
    }
  }
}
```

#### **1.2 User Login**

```http
POST http://103.208.137.57/alulz/api/v1/auth/login
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025

{
  "email": "john.doe@example.com",
  "password": "password123"
}
```

#### **1.3 Refresh Token**

```http
POST http://103.208.137.57/alulz/api/v1/auth/refresh
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025

{
  "refresh_token": "your_refresh_token_here"
}
```

#### **1.4 Logout**

```http
POST http://103.208.137.57/alulz/api/v1/auth/logout
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <your_access_token>

{
  "refresh_token": "your_refresh_token_here"
}
```

### **Phase 2: Customer Operations**

#### **2.1 Get User Profile**

```http
GET http://103.208.137.57/alulz/api/v1/user/profile
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <your_access_token>
```

#### **2.2 Update User Profile**

```http
PUT http://103.208.137.57/alulz/api/v1/user/profile
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <your_access_token>

{
  "name": "John Updated Doe",
  "phone": "081234567891"
}
```

#### **2.3 Get Nearby Parking Areas**

```http
GET http://103.208.137.57/alulz/api/v1/parking/locations?latitude=-2.9881&longitude=104.7591&radius=1000
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <your_access_token>
```

#### **2.4 Check-in to Parking Area**

```http
POST http://103.208.137.57/alulz/api/v1/parking/checkin
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <your_access_token>

{
  "area_id": 1,
  "vehicle_plate": "B1234ABC"
}
```

#### **2.5 Check-out from Parking Area**

```http
POST http://103.208.137.57/alulz/api/v1/parking/checkout
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <your_access_token>

{
  "session_id": 1
}
```

#### **2.6 Get Active Parking Session**

```http
GET http://103.208.137.57/alulz/api/v1/parking/session
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <your_access_token>
```

### **Phase 3: Jukir Operations**

#### **3.1 Register as Jukir**

```http
POST http://103.208.137.57/alulz/api/v1/auth/register
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025

{
  "name": "Jukir Name",
  "email": "jukir@example.com",
  "password": "password123",
  "phone": "081234567892",
  "role": "jukir"
}
```

#### **3.2 Login as Jukir**

```http
POST http://103.208.137.57/alulz/api/v1/auth/login
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025

{
  "email": "jukir@example.com",
  "password": "password123"
}
```

#### **3.3 Get Jukir Dashboard**

```http
GET http://103.208.137.57/alulz/api/v1/jukir/dashboard
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <jukir_access_token>
```

#### **3.4 Get Jukir Payments**

```http
GET http://103.208.137.57/alulz/api/v1/jukir/payments
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <jukir_access_token>
```

#### **3.5 Confirm Payment**

```http
POST http://103.208.137.57/alulz/api/v1/jukir/confirm-payment
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <jukir_access_token>

{
  "payment_id": 1
}
```

#### **3.6 Generate QR Code**

```http
GET http://103.208.137.57/alulz/api/v1/jukir/qr-code
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <jukir_access_token>
```

#### **3.7 Get Daily Report**

```http
GET http://103.208.137.57/alulz/api/v1/jukir/daily-report?date=2025-10-25
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <jukir_access_token>
```

#### **3.8 Manual Check-in**

```http
POST http://103.208.137.57/alulz/api/v1/jukir/manual-checkin
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <jukir_access_token>

{
  "customer_phone": "081234567890",
  "vehicle_plate": "B1234ABC",
  "area_id": 1
}
```

#### **3.9 Manual Check-out**

```http
POST http://103.208.137.57/alulz/api/v1/jukir/manual-checkout
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <jukir_access_token>

{
  "session_id": 1
}
```

#### **3.10 Connect to Real-time Events (SSE)**

```http
GET http://103.208.137.57/alulz/api/v1/jukir/events
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <jukir_access_token>
Accept: text/event-stream
```

### **Phase 4: Admin Operations**

#### **4.1 Register as Admin**

```http
POST http://103.208.137.57/alulz/api/v1/auth/register
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025

{
  "name": "Admin Name",
  "email": "admin@example.com",
  "password": "password123",
  "phone": "081234567893",
  "role": "admin"
}
```

#### **4.2 Login as Admin**

```http
POST http://103.208.137.57/alulz/api/v1/auth/login
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025

{
  "email": "admin@example.com",
  "password": "password123"
}
```

#### **4.3 Get Admin Overview**

```http
GET http://103.208.137.57/alulz/api/v1/admin/overview
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <admin_access_token>
```

#### **4.4 Get All Jukirs**

```http
GET http://103.208.137.57/alulz/api/v1/admin/jukirs
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <admin_access_token>
```

#### **4.5 Create New Jukir**

```http
POST http://103.208.137.57/alulz/api/v1/admin/jukirs
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <admin_access_token>

{
  "name": "New Jukir",
  "email": "newjukir@example.com",
  "password": "password123",
  "phone": "081234567894"
}
```

#### **4.6 Update Jukir Status**

```http
PUT http://103.208.137.57/alulz/api/v1/admin/jukirs/1/status
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <admin_access_token>

{
  "status": "active"
}
```

#### **4.7 Get Reports**

```http
GET http://103.208.137.57/alulz/api/v1/admin/reports?start_date=2025-10-01&end_date=2025-10-31
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <admin_access_token>
```

#### **4.8 Get All Sessions**

```http
GET http://103.208.137.57/alulz/api/v1/admin/sessions?page=1&limit=10
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <admin_access_token>
```

#### **4.9 Create Parking Area**

```http
POST http://103.208.137.57/alulz/api/v1/admin/areas
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <admin_access_token>

{
  "name": "Mall Palembang",
  "address": "Jl. Sudirman No. 1, Palembang",
  "latitude": -2.9881,
  "longitude": 104.7591,
  "capacity": 100,
  "hourly_rate": 2000,
  "description": "Parking area at Mall Palembang"
}
```

#### **4.10 Update Parking Area**

```http
PUT http://103.208.137.57/alulz/api/v1/admin/areas/1
Content-Type: application/json
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <admin_access_token>

{
  "name": "Mall Palembang Updated",
  "capacity": 150,
  "hourly_rate": 2500
}
```

#### **4.11 Get SSE Status**

```http
GET http://103.208.137.57/alulz/api/v1/admin/sse-status
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <admin_access_token>
```

## 🔄 **Complete Business Flow Testing**

### **Scenario 1: Customer Parking Flow**

1. **Register Customer** → Get access token
2. **Get Nearby Areas** → Find available parking
3. **Check-in** → Start parking session
4. **Get Active Session** → Verify session status
5. **Check-out** → End parking session

### **Scenario 2: Jukir Management Flow**

1. **Register Jukir** → Get Jukir access token
2. **Get Dashboard** → View current status
3. **Manual Check-in** → Register customer manually
4. **Confirm Payment** → Process payment
5. **Get Daily Report** → View daily earnings

### **Scenario 3: Admin Management Flow**

1. **Register Admin** → Get admin access token
2. **Get Overview** → View system statistics
3. **Create Parking Area** → Add new parking location
4. **Create Jukir** → Add new Jukir user
5. **Get Reports** → View system reports

### **Scenario 4: Real-time Notifications**

1. **Jukir connects to SSE** → `GET /api/v1/jukir/events`
2. **Customer checks out** → Triggers SSE event
3. **Jukir receives notification** → Real-time update

## 📊 **Expected Response Formats**

### **Success Response**

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {
    // Response data here
  }
}
```

### **Error Response**

```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error information"
}
```

### **Pagination Response**

```json
{
  "success": true,
  "message": "Data retrieved successfully",
  "data": {
    "items": [],
    "pagination": {
      "page": 1,
      "limit": 10,
      "total": 100,
      "total_pages": 10
    }
  }
}
```

## 🧪 **Testing Checklist**

### **Authentication Tests**

- [ ] User registration (customer, jukir, admin)
- [ ] User login with valid credentials
- [ ] Token refresh functionality
- [ ] Logout functionality
- [ ] Invalid credentials handling

### **Customer Tests**

- [ ] Profile management
- [ ] Nearby areas search
- [ ] Check-in process
- [ ] Check-out process
- [ ] Active session retrieval

### **Jukir Tests**

- [ ] Dashboard access
- [ ] Payment management
- [ ] QR code generation
- [ ] Manual operations
- [ ] Daily reports
- [ ] SSE connection

### **Admin Tests**

- [ ] System overview
- [ ] User management
- [ ] Parking area management
- [ ] Report generation
- [ ] SSE monitoring

### **Integration Tests**

- [ ] Complete customer flow
- [ ] Complete Jukir flow
- [ ] Complete admin flow
- [ ] Real-time notifications
- [ ] Error handling

## 🚨 **Common Issues & Solutions**

### **401 Unauthorized**

- Check if `X-API-Key` header is included
- Verify JWT token is valid and not expired
- Ensure proper `Authorization: Bearer <token>` format

### **403 Forbidden**

- Verify user has correct role permissions
- Check if user account is active
- Ensure proper role-based access

### **404 Not Found**

- Verify correct API endpoint URL
- Check if resource exists in database
- Ensure proper HTTP method

### **500 Internal Server Error**

- Check server logs for detailed error
- Verify database connectivity
- Check Redis connection status

## 📝 **Postman Collection Setup**

### **Environment Variables**

Create a Postman environment with:

```
base_url: http://103.208.137.57/alulz
api_key: be-parkir-api-key-2025
access_token: (will be set dynamically)
refresh_token: (will be set dynamically)
customer_token: (for customer operations)
jukir_token: (for Jukir operations)
admin_token: (for admin operations)
```

### **Pre-request Scripts**

Add to collection pre-request script:

```javascript
// Set API key for all requests
pm.request.headers.add({
  key: "X-API-Key",
  value: pm.environment.get("api_key"),
});
```

### **Test Scripts**

Add to authentication requests:

```javascript
// Save tokens after login/register
if (pm.response.code === 200 || pm.response.code === 201) {
  const response = pm.response.json();
  if (response.data.access_token) {
    pm.environment.set("access_token", response.data.access_token);
  }
  if (response.data.refresh_token) {
    pm.environment.set("refresh_token", response.data.refresh_token);
  }
}
```

## 🎯 **Performance Testing**

### **Load Testing Endpoints**

- Health check: `GET /health`
- Authentication: `POST /api/v1/auth/login`
- Nearby areas: `GET /api/v1/parking/locations`
- Dashboard: `GET /api/v1/jukir/dashboard`

### **Expected Response Times**

- Health check: < 100ms
- Authentication: < 500ms
- Database queries: < 1000ms
- SSE connections: Real-time

---

## 📞 **Support Information**

- **API Documentation**: `http://103.208.137.57/alulz/swagger/index.html`
- **Health Check**: `http://103.208.137.57/alulz/health`
- **Base URL**: `http://103.208.137.57/alulz`

This guide provides comprehensive testing coverage for all Be-Parkir API endpoints and business flows. Use it to systematically test the entire system functionality.
