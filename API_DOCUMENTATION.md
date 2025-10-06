# Be-Parkir API Documentation for Postman

## 🚀 **API Overview**

The Be-Parkir API is a digital parking system for Palembang with anonymous parking, manual record management, and role-based access control.

**Base URL**: `http://localhost:8080/api/v1`

## 📋 **Authentication**

The API uses multiple layers of authentication:

### **1. API Key Authentication (Required for all endpoints)**

All API endpoints require an API key in the request header:

```
X-API-Key: be-parkir-api-key-2025
```

### **2. JWT Authentication (For protected endpoints)**

Protected endpoints also require JWT-based authentication. Include the token in the Authorization header:

```
Authorization: Bearer <your-jwt-token>
```

### **3. CORS Support**

The API supports Cross-Origin Resource Sharing (CORS) for web applications:

**Allowed Origins:**

- `http://localhost:3000`
- `http://localhost:3001`
- `http://localhost:8080`
- `https://parkir.palembang.go.id`

**Allowed Headers:**

- `Origin`, `Content-Type`, `Accept`
- `Authorization`, `X-Requested-With`
- `X-API-Key`, `X-Client-Version`, `X-Device-ID`

## 🎯 **Business Process Testing Guide**

### **Phase 1: Setup (Admin & Jukir Registration)**

#### 1. **Register Admin User**

```http
POST /api/v1/auth/register
X-API-Key: be-parkir-api-key-2025
Content-Type: application/json

{
  "name": "Admin System",
  "email": "admin@parkir.com",
  "phone": "08123456789",
  "password": "admin123",
  "role": "admin"
}
```

#### 2. **Login as Admin**

```http
POST /api/v1/auth/login
X-API-Key: be-parkir-api-key-2025
Content-Type: application/json

{
  "email": "admin@parkir.com",
  "password": "admin123"
}
```

**Response**: Save the `access_token` for admin operations.

#### 3. **Create Parking Area (Admin)**

```http
POST /api/v1/admin/areas
X-API-Key: be-parkir-api-key-2025
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "name": "Parkir Mall Palembang",
  "address": "Jl. Sudirman No. 1, Palembang",
  "latitude": -2.9761,
  "longitude": 104.7754,
  "hourly_rate": 2000.00
}
```

#### 4. **Register Jukir User**

```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "name": "Jukir Sample",
  "email": "jukir@parkir.com",
  "phone": "08123456790",
  "password": "jukir123",
  "role": "jukir"
}
```

#### 5. **Create Jukir Account (Admin)**

```http
POST /api/v1/admin/jukirs
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "user_id": 2,
  "jukir_code": "JUK001",
  "area_id": 1
}
```

#### 6. **Activate Jukir (Admin)**

```http
PUT /api/v1/admin/jukirs/1/status
Authorization: Bearer <admin-token>
Content-Type: application/json

{
  "status": "active"
}
```

#### 7. **Login as Jukir**

```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "email": "jukir@parkir.com",
  "password": "jukir123"
}
```

**Response**: Save the `access_token` for jukir operations.

### **Phase 2: Anonymous Parking (QR-Based)**

#### 8. **Get Nearby Parking Areas (Anonymous)**

```http
GET /api/v1/parking/locations?latitude=-2.9761&longitude=104.7754&radius=1.0
X-API-Key: be-parkir-api-key-2025
```

#### 9. **QR-Based Check-in (Anonymous)**

```http
POST /api/v1/parking/checkin
X-API-Key: be-parkir-api-key-2025
Content-Type: application/json

{
  "qr_token": "QR_JUK001_20250101120000",
  "latitude": -2.9761,
  "longitude": 104.7754,
  "vehicle_type": "mobil"
}
```

**Note**: No license plate required for QR-based sessions.

#### 10. **Get Active Session (Anonymous)**

```http
GET /api/v1/parking/active?qr_token=QR_JUK001_20250101120000
```

#### 11. **QR-Based Check-out (Anonymous)**

```http
POST /api/v1/parking/checkout
Content-Type: application/json

{
  "qr_token": "QR_JUK001_20250101120000",
  "latitude": -2.9761,
  "longitude": 104.7754
}
```

### **Phase 3: Manual Records (Jukir)**

#### 12. **Manual Check-in (Jukir)**

```http
POST /api/v1/jukir/manual-checkin
Authorization: Bearer <jukir-token>
Content-Type: application/json

{
  "plat_nomor": "B1234ABC",
  "vehicle_type": "mobil",
  "waktu_masuk": "2025-01-01T12:00:00Z"
}
```

#### 13. **Manual Check-out (Jukir)**

```http
POST /api/v1/jukir/manual-checkout
Authorization: Bearer <jukir-token>
Content-Type: application/json

{
  "session_id": 2,
  "waktu_keluar": "2025-01-01T14:30:00Z"
}
```

### **Phase 4: Payment Processing (Jukir)**

#### 14. **View Pending Payments (Jukir)**

```http
GET /api/v1/jukir/pending-payments
Authorization: Bearer <jukir-token>
```

#### 15. **Confirm Payment (Jukir)**

```http
POST /api/v1/jukir/confirm-payment
Authorization: Bearer <jukir-token>
Content-Type: application/json

{
  "session_id": 1,
  "payment_method": "cash"
}
```

### **Phase 5: Dashboard & Reports**

#### 16. **Jukir Dashboard**

```http
GET /api/v1/jukir/dashboard
Authorization: Bearer <jukir-token>
```

#### 17. **Jukir Daily Report**

```http
GET /api/v1/jukir/daily-report?date=2025-01-01
Authorization: Bearer <jukir-token>
```

#### 18. **Admin Overview**

```http
GET /api/v1/admin/overview
Authorization: Bearer <admin-token>
```

#### 19. **Admin Reports**

```http
GET /api/v1/admin/reports?start_date=2025-01-01&end_date=2025-01-01
Authorization: Bearer <admin-token>
```

## 📊 **Postman Collection Structure**

### **Environment Variables**

Create a Postman environment with these variables:

```json
{
  "base_url": "http://localhost:8080/api/v1",
  "api_key": "be-parkir-api-key-2025",
  "admin_token": "",
  "jukir_token": "",
  "area_id": "",
  "jukir_id": "",
  "session_id": ""
}
```

### **Collection Folders**

#### **1. Authentication**

- `POST /auth/register` (Admin)
- `POST /auth/register` (Jukir)
- `POST /auth/login` (Admin)
- `POST /auth/login` (Jukir)
- `POST /auth/refresh`
- `POST /auth/logout`

#### **2. Admin Management**

- `GET /admin/overview`
- `POST /admin/areas`
- `GET /admin/jukirs`
- `POST /admin/jukirs`
- `PUT /admin/jukirs/{id}/status`
- `GET /admin/reports`
- `GET /admin/sessions`

#### **3. Anonymous Parking**

- `GET /parking/locations`
- `POST /parking/checkin`
- `POST /parking/checkout`
- `GET /parking/active`
- `GET /parking/history`

#### **4. Jukir Operations**

- `GET /jukir/dashboard`
- `GET /jukir/pending-payments`
- `GET /jukir/active-sessions`
- `POST /jukir/confirm-payment`
- `GET /jukir/qr-code`
- `GET /jukir/daily-report`
- `POST /jukir/manual-checkin`
- `POST /jukir/manual-checkout`

#### **5. User Management**

- `GET /profile`
- `PUT /profile`

## 🧪 **Test Scenarios**

### **Scenario 1: Complete QR-Based Parking Flow**

1. Register admin and jukir
2. Create parking area
3. Create and activate jukir
4. Anonymous QR check-in
5. Anonymous QR check-out
6. Jukir confirm payment

### **Scenario 2: Manual Record Flow**

1. Jukir manual check-in
2. Jukir manual check-out
3. Jukir confirm payment

### **Scenario 3: Admin Management**

1. View system overview
2. Manage jukirs
3. Generate reports
4. View all sessions

## 📝 **Response Examples**

### **Success Response Format**

```json
{
  "success": true,
  "message": "Operation successful",
  "data": {
    // Response data
  },
  "meta": {
    "pagination": {
      "limit": 10,
      "offset": 0,
      "total": 100
    }
  }
}
```

### **Error Response Format**

```json
{
  "success": false,
  "message": "Error description",
  "error": "Detailed error information"
}
```

## 🔧 **Testing Tips**

1. **API Key Required**: All requests must include `X-API-Key` header
2. **Use Environment Variables** for tokens and IDs
3. **Test Anonymous Endpoints** without JWT authentication (but with API key)
4. **Verify GPS Coordinates** are within 50m of parking area
5. **Check Vehicle Types** are 'mobil' or 'motor'
6. **Test Manual Records** require license plates
7. **Configurable Pricing**: Each parking area has its own hourly rate
8. **CORS Support**: Frontend applications can make cross-origin requests

## 🚨 **Common Test Cases**

### **GPS Validation Test**

```json
{
  "qr_token": "QR_JUK001_20250101120000",
  "latitude": -2.9761,
  "longitude": 104.7754,
  "vehicle_type": "mobil"
}
```

### **Vehicle Type Validation**

```json
{
  "vehicle_type": "mobil" // or "motor"
}
```

### **Manual Record Test**

```json
{
  "plat_nomor": "B1234ABC",
  "vehicle_type": "mobil",
  "waktu_masuk": "2025-01-01T12:00:00Z"
}
```

## 🧪 **Test Results Summary**

### ✅ **All Business Processes Tested Successfully**

**Date**: October 6, 2025  
**Status**: All endpoints working correctly

#### **Authentication Flow** ✅

- ✅ Admin registration and login
- ✅ Jukir registration and login
- ✅ JWT token generation and validation
- ✅ Role-based access control

#### **Admin Management Flow** ✅

- ✅ Create parking areas
- ✅ Create and activate Jukir accounts
- ✅ View system overview and statistics
- ✅ Generate reports

#### **Anonymous Parking Flow** ✅

- ✅ Get nearby parking areas (GPS-based)
- ✅ QR-based check-in (no license plate required)
- ✅ QR-based check-out with area-specific pricing
- ✅ Session management

#### **Jukir Operations Flow** ✅

- ✅ Manual check-in with license plate
- ✅ Manual check-out with time calculation
- ✅ View pending payments
- ✅ Confirm payments
- ✅ Dashboard with statistics

#### **Payment Processing** ✅

- ✅ Area-based cost calculation
- ✅ Payment confirmation
- ✅ Revenue tracking

### 📊 **Test Data Generated**

- **Users**: 2 (1 Admin, 1 Jukir)
- **Parking Areas**: 1 (Parkir Mall Palembang)
- **Jukir Accounts**: 1 (JUK001, Active)
- **Parking Sessions**: 2 (1 QR-based, 1 Manual)
- **Payments**: 1 Confirmed (IDR 2000)

### 🎯 **Key Features Verified**

- ✅ Anonymous parking (no user account required)
- ✅ Vehicle type support (mobil/motor)
- ✅ Configurable pricing system (area-specific rates)
- ✅ Manual record management by Jukirs
- ✅ GPS-based area detection
- ✅ Role-based access control
- ✅ Real-time statistics and reporting

## 📁 **Postman Files**

### **Collection File**

- `Be-Parkir-API.postman_collection.json` - Complete API collection with all endpoints

### **Environment File**

- `Be-Parkir-Environment.postman_environment.json` - Environment variables for testing

### **Import Instructions**

1. Open Postman
2. Click "Import" button
3. Select both JSON files
4. Set the environment to "Be-Parkir Environment"
5. Run the collection in sequence

## 🔐 **Security & Middleware Features**

### **API Key Authentication**

- **Purpose**: First layer of security for all API endpoints
- **Header**: `X-API-Key: be-parkir-api-key-2025`
- **Configuration**: Set via `API_KEY` environment variable
- **Validation**: All requests must include valid API key

### **CORS (Cross-Origin Resource Sharing)**

- **Purpose**: Enable web applications to access the API
- **Preflight**: Automatic OPTIONS request handling
- **Headers**: Configurable allowed origins and headers
- **Credentials**: Support for authenticated requests

### **JWT Authentication**

- **Purpose**: User-specific authentication for protected endpoints
- **Header**: `Authorization: Bearer <token>`
- **Roles**: Admin, Jukir role-based access control
- **Expiry**: Configurable token expiration

### **Environment Configuration**

```env
# API Key Configuration
API_KEY=be-parkir-api-key-2025
API_KEY_REQUIRED=true
API_KEY_HEADER=X-API-Key

# CORS Configuration
CORS_ALLOW_ORIGINS=http://localhost:3000,http://localhost:3001,https://parkir.palembang.go.id
CORS_ALLOW_CREDENTIALS=true
CORS_MAX_AGE=86400

# JWT Configuration
JWT_SECRET=your-super-secret-jwt-key
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d
```

## 🚀 **Production Deployment Notes**

### **Security Checklist**

- ✅ API key authentication enabled
- ✅ CORS properly configured for production domains
- ✅ JWT secrets are secure and unique
- ✅ Database connections are encrypted
- ✅ Environment variables are properly set
- ✅ Rate limiting can be added for additional security

### **Frontend Integration**

```javascript
// Example frontend API call
const response = await fetch("http://localhost:8080/api/v1/parking/locations", {
  method: "GET",
  headers: {
    "X-API-Key": "be-parkir-api-key-2025",
    "Content-Type": "application/json",
    Authorization: "Bearer " + userToken, // For authenticated endpoints
  },
});
```

This documentation provides a complete testing guide for all business processes in the Be-Parkir system! 🎉
