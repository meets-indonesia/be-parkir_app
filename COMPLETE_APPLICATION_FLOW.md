# Be-Parkir - Complete Application Flow Documentation

## 📋 Table of Contents

1. [Admin Flow](#1-admin-flow)
2. [Mobile User Flow](#2-mobile-user-flow)
3. [Jukir Flow](#3-jukir-flow)
4. [API Endpoints Summary](#4-api-endpoints-summary)

---

## 1. ADMIN FLOW

### 1.1 Registration & Login

#### Step 1: Admin Register

**Endpoint**: `POST /api/v1/auth/register`

**Request Body**:

```json
{
  "name": "Admin Parkir",
  "email": "admin@parkir.com",
  "phone": "081234567890",
  "password": "admin123"
}
```

**Response**:

```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "id": 1,
    "name": "Admin Parkir",
    "email": "admin@parkir.com",
    "role": "admin",
    "status": "active"
  }
}
```

**Process**:

- Admin register dengan email, nama, phone, password
- User dibuat dengan role "admin" dan status "active"
- Password di-hash menggunakan bcrypt

---

#### Step 2: Admin Login

**Endpoint**: `POST /api/v1/auth/login`

**Request Body**:

```json
{
  "email": "admin@parkir.com",
  "password": "admin123"
}
```

**Response**:

```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "refresh_token_string",
    "user": {
      "id": 1,
      "name": "Admin Parkir",
      "email": "admin@parkir.com",
      "role": "admin"
    }
  }
}
```

**Process**:

- Validasi email dan password
- Generate JWT access token (exp: 1 hour)
- Generate refresh token dan simpan di Redis (exp: 7 days)
- Return tokens dan user info

---

### 1.2 Parking Area Management

#### Step 3: Create Parking Area

**Endpoint**: `POST /api/v1/admin/areas`
**Auth**: Admin token required

**Request Body**:

```json
{
  "name": "Parkir Mall Palembang",
  "address": "Jl. Sudirman No. 1, Palembang",
  "latitude": -2.9761,
  "longitude": 104.7754,
  "hourly_rate": 2000,
  "max_mobil": 50,
  "max_motor": 100,
  "status_operasional": "buka",
  "jenis_area": "outdoor"
}
```

**Response**:

```json
{
  "success": true,
  "message": "Parking area created successfully",
  "data": {
    "id": 1,
    "name": "Parkir Mall Palembang",
    "address": "Jl. Sudirman No. 1, Palembang",
    "latitude": -2.9761,
    "longitude": 104.7754,
    "hourly_rate": 2000,
    "max_mobil": 50,
    "max_motor": 100,
    "status_operasional": "buka",
    "jenis_area": " melanoma",
    "status": "active",
    "created_at": "2025-01-20T10:00:00Z"
  }
}
```

**Process**:

- Create area dengan kapasitas mobil dan motor
- Set status operasional (buka/tutup/maintenance)
- Set jenis area (indoor/outdoor/mix)

---

### 1.3 Jukir Management

#### Step 4: Create Jukir Account

**Endpoint**: `POST /api/v1/admin/jukirs`
**Auth**: Admin token required

**Request Body**:

```json
{
  "name": "Budi Santoso",
  "area_id": 1,
  "status": "active"
}
```

**Response**:

```json
{
  "success": true,
  "message": "Jukir created successfully",
  "data": {
    "jukir": {
      "id": 1,
      "user_id": 10,
      "jukir_code": "MAL1430",
      "area_id": 1,
      "qr_token": "QR_MAL1430_20250120143000",
      "status": "active"
    },
    "username": "mal1430",
    "password": "1234"
  }
}
```

**Process**:

1. Generate jukir code (3 huruf area + HHMM)
2. Generate username (lowercase jukir code)
3. Generate password (4 digit random)
4. Create user account with role "jukir"
5. Create jukir record with QR token
6. Return username & password (plain text untuk diberikan ke jukir)

---

#### Step 5: List All Jukirs

**Endpoint**: `GET /api/v1/admin/jukirs/list`
**Auth**: Admin token required

**Query Parameters**:

- `include_revenue` (optional): true/false
- `status` (optional): active/inactive/pending
- `vehicle_type` (optional): mobil/motor
- `date_range` (optional): hari*ini/minggu*現在的 /bulan_ini

**Example Request**:

```bash
curl -X GET "http://localhost:8080/api/v1/admin/jukirs/list?include_revenue=true&status=active&date_range=minggu_ini" \
  -H "Authorization: Bearer ACCESS_TOKEN" \
  -H "X-API-Key: API_KEY"
```

**Response** (with revenue):

```json
{
  "success": true,
  "message": "Jukirs list retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "Budi Santoso",
      "status": "active",
      "area": {
        "id": 1,
        "name": "Parkir Mall Palembang"
      },
      "jukir_code": "MAL1430",
      "actual_revenue": 350000,
      "estimated_revenue": 1050000,
      "total_revenue": 1400000,
      "date": "2025-01-14"
    }
  ]
}
```

---

#### Step 6: Get Jukir Detail

**Endpoint**: `GET /api/v1/admin/jukirs/:id`
**Auth**: Admin token required

**Response**:

```json
{
  "success": true,
  "message": "Jukir retrieved successfully",
  "data": {
    "id": 1,
    "name": "Budi Santoso",
    "status": "active",
    "jukir_code": "MAL1430",
    "qr_token": "QR_MAL1430_20250120143000",
    "user": {
      "id": 10,
      "name": "Budi Santoso",
      "email": "mal1430"
    },
    "area": {
      "id": 1,
      "name": "Parkir Mall Palembang",
      "address": "Jl. Sudirman No. 1, Palembang"
    },
    "created_at": "2025-01-20T14:30:00Z",
    "updated_at": "2025-01-20T14:30:00Z"
  }
}
```

---

### 1.4 Dashboard Management

#### Step 7: Get Admin Overview

**Endpoint**: `GET /api/v1/admin/overview`
**Auth**: Admin token required

**Query Parameters**:

- `vehicle_type` (optional): mobil/motor
- `date_range` (optional): hari_ini/minggu_ini/bulan_ini/tahun_ini

**Response**:

```json
{
  "success": true,
  "message": "Overview data retrieved successfully",
  "data": {
    "total_users": 150,
    "total_jukirs": 25,
    "total_areas": 5,
    "today_sessions": 45,
    "vehicles_in": 45,
    "vehicles_out": 32,
    "vehicles_by_type": {
      "mobil": { "in": 20, "out": 15 },
      "motor": { "in": 25, "out": 17 }
    },
    "active_sessions": 13,
    "pending_payments": 8,
    "today_revenue": 350000,
    "estimated_revenue": 425000,
    "jukir_status": {
      "active": 18,
      "inactive": 7
    },
    "chart_data": [
      {
        "period": "Min",
        "date": "2025-01-14",
        "actual_revenue": 125000,
        "estimated_revenue": 150000
      },
      ...
    ]
  },
  "filter": {
    "vehicle_type": "mobil",
    "date_range": "hari_ini"
  }
}
```

---

#### Step 8: Get Vehicle Statistics

**Endpoint**: `GET /api/v1/admin/statistics/vehicles`
**Auth**: Admin token required

**Query Parameters**:

- `vehicle_type` (optional): mobil/motor
- `date_range` (optional): hari_ini/minggu_ini/bulan_ini

**Response**:

```json
{
  "success": true,
  "message": "Vehicle statistics retrieved successfully",
  "data": {
    "total_in": 250,
    "total_out": 180,
    "vehicles_by_type": {
      "mobil": { "in": 120, "out": 90 },
      "motor": { "in": 130, "out": 90 }
    },
    "date_range": "minggu_ini"
  }
}
```

---

#### Step 9: Get Total Revenue

**Endpoint**: `GET /api/v1/admin/revenue/total`
**Auth**: Admin token required

**Query Parameters**:

- `vehicle_type` (optional): mobil/motor
- `date_range` (optional): hari_ini/minggu_ini/bulan_ini

**Response**:

```json
{
  "success": true,
  "message": "Total revenue retrieved successfully",
  "data": {
    "actual_revenue": 8500000,
    "estimated_revenue": 12000000,
    "total_revenue": 20500000,
    "date_range": "bulan_ini"
  }
}
```

---

#### Step 10: Get Jukir Statistics

**Endpoint**: `GET /api/v1/admin/statistics/jukirs`
**Auth**: Admin token required

**Response**:

```json
{
  "success": true,
  "message": "Jukir statistics retrieved successfully",
  "data": {
    "total": 25,
    "active": 18,
    "inactive": 7
  }
}
```

---

#### Step 11: Get Parking Area Statistics

**Endpoint**: `GET /api/v1/admin/statistics/areas`
**Auth**: Admin token required

**Response**:

```json
{
  "success": true,
  "message": "Parking area statistics retrieved successfully",
  "data": {
    "total": 10,
    "active": 7,
    "inactive": 2,
    "maintenance": 1
  }
}
```

---

#### Step 12: Get Chart Data

**Endpoint**: `GET /api/v1/admin/chart/data`
**Auth**: Admin token required

**Query Parameters**:

- `vehicle_type` (optional): mobil/motor
- `date_range` (optional): minggu_ini/bulan_ini

**Response**:

```json
{
  "success": true,
  "message": "Chart data retrieved successfully",
  "data": [
    {
      "period": "Minggu 1",
      "date": "2025-01-08",
      "actual_revenue": 1250000,
      "estimated_revenue": 1500000
    },
    ...
  ]
}
```

---

### 1.5 Manual Revenue Management

#### Step 13: Add Manual Revenue

**Endpoint**: `POST /api/v1/admin/jukirs/manual-revenue`
**Auth**: Admin token required

**Request Body**:

```json
{
  "jukir_id": 1,
  "amount": 50000,
  "date": "2025-01-20",
  "notes": "Penambahan pendapatan manual dari parkir tambahan"
}
```

**Response**:

```json
{
  "success": true,
  "message": "Manual revenue added successfully",
  "data": {
    "id": 1,
    "jukir_name": "Budi Santoso",
    "actual_revenue": 50000,
    "estimated_revenue": 0,
    "total_revenue": 50000,
    "date": "2025-01-20"
  }
}
```

---

---

## 2. MOBILE USER FLOW

### 2.1 User Registration & Login

#### Step 1: User Register

**Endpoint**: `POST /api/v1/auth/register`

**Request Body**:

```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "phone": "081234567891",
  "password": "password123"
}
```

**Response**:

```json
{
  "success": true,
  "message": "User registered successfully",
  "data": {
    "id": 5,
    "name": "John Doe",
    "email": "john@example.com",
    "role": "user",
    "status": "active"
  }
}
```

---

#### Step 2: User Login

**Endpoint**: `POST /api/v1/auth/login`

**Request Body**:

```json
{
  "email": "john@example.com",
  "password": "password123"
}
```

**Response**:

```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "refresh_token_string",
   Status": "user"
  }
}
```

---

### 2.2 Parking Session Flow

#### Step 3: Find Nearby Parking Areas

**Endpoint**: `GET /api/v1/parking/locations`
**Auth**: No authentication required

**Query Parameters**:

- `latitude` (required)
- `longitude` (required)
- `radius` (optional, default: 1000 meter)

**Example Request**:

```bash
curl -X GET "http://localhost:8080/api/v1/parking/locations?latitude=-2.9761&longitude=104.7754&radius=500" \
  -H "X-API-Key: API_KEY"
```

**Response**:

```json
{
  "success": true,
  "message": "Nearby areas retrieved successfully",
  "data": [
    {
      "id": 1,
      "name": "Parkir Mall Palembang",
      "address": "Jl. Sudirman No. 1, Palembang",
      "distance": 250.5,
      "hourly_rate": 2000,
      "available_mobil": 35,
      "available_motor": 75
    }
  ]
}
```

---

#### Step 4: Check-In (Parking Start)

**Endpoint**: `POST /api/v1/parking/checkin`
**Auth**: No authentication required (Anonymous)

**Request Body**:

```json
{
  "qr_token": "QR_MAL1430_20250120143000",
  "vehicle_type": "mobil",
  "plat_nomor": "BG1234XYZ",
  "latitude": -2.9761,
  "longitude": 104.7754
}
```

**Response**:

```json
{
  "success": true,
  "message": "Check-in successful",
  "data": {
    "session_id": 123,
    "checkin_time": "2025-01-20T14:30:00Z",
    "qr_token": "QR_MAL1430_20250120143000",
    "area_name": "Parkir Mall Palembang",
    "status": "active"
  }
}
```

**Process**:

1. Validasi QR token
2. Cari jukir dan area berdasarkan QR token
3. Create parking session dengan status "active"
4. Generate session ID
5. Set waktu check-in
6. Return session details

---

#### Step 5: Get Active Session

**Endpoint**: `GET /api/v1/parking/active`
**Auth**: No authentication required

**Query Parameters**:

- `qr_token` (bol)

**Example Request**:

```bash
curl -X GET "http://localhost:8080/api/v1/parking/active?qr_token=QR_MAL1430_20250120143000" \
  -H "X-API-Key: API_KEY"
```

**Response**:

```json
{
  "success": true,
  "message": "Active session retrieved successfully",
  "data梵丸t": {
    "session_id": 123,
    "checkin_time": "2025-01-20T14:30:00Z",
    "duration_minutes": 45,
    "estimated_cost": 1500,
    "area_name": "Parkir Mall Palembang",
    "hourly_rate": 2000
  }
}
```

---

#### Step 6: Check-Out (End Parking)

**Endpoint**: `POST /api/v1/parking/checkout`
**Auth**: No authentication required

**Request Body**:

```json
{
  "qr_token": "QR_MAL1430_20250120143000",
  "session_id": 123,
  "plat_nomor": "BG1234XYZ",
  "latitude": -2.9761,
  "longitude": 104.7754
}
```

**Response**:

```json
{
  "success": true,
  "message": "Check-out successful",
  "data": {
    "session_id": 123,
    "checkin_time": "2025-ります20143000Z",
    "checkout_time": "2025-01-20T15:30:00Z",
    "duration_minutes": 60,
    "total_cost": 2000,
    "payment_status": "pending",
    "area_name": "Parkir Mall Palembang"
  }
}
```

**Process**:

1. Identify session by session_id atau plat_nomor
2. Validasi session exists dan masih active
3. Set waktu checkout
4. Calculate total cost (duration × hourly_rate)
5. Update session status ke "pending_payment"
6. Return invoice details

---

#### Step 7: View Parking History

**Endpoint**: `GET /api/v1/parking/history`
**Auth**: No authentication required

**Query Parameters**:

- `plat_nomor` (optional)
- `session_id` (optional)
- `limit` (optional, default: 10)
- `offset` (optional, default: 0)

**Example Request**:

```bash
curl -X GET "http://localhost:8080/api/v1/parking/history?plat_nomor=BG1234XYZ&limit=10&offset=0" \
  -H "X-API-Key: API_KEY"
```

**Response**:

```json
{
  "success": true,
  "message": "Parking history retrieved successfully",
  "data": {
    "sessions": [
      {
        "session_id": 123,
        "checkin_time": "2025-01-20T14:30:00Z",
        "checkout_time": "2025-01-20T15:30:00Z",
        "duration_minutes": 60,
        "total_cost": 2000,
        "payment_status": "paid",
        "area_name": "Parkir Mall Palembang",
        "vehicle_type": "mobil"
      }
    ],
    "count": 1
  },
  "meta": {
    "pagination": {
      "limit": 10,
      "offset": 0,
      "total": 1
    }
  }
}
```

---

### 2.3 User Profile Management

#### Step 8: Get User Profile

**Endpoint**: `GET /api/v1/profile`
**Auth**: User token required

**Response**:

```json
{
  "success": true,
  "message": "Profile retrieved successfully",
  "data": {
    "id": 5,
    "name": "John Doe",
    "email": "john@example.com",
    "phone": "081234567891",
    "role": "user",
    "status": "active"
  }
}
```

---

#### Step 9: Update User Profile

**Endpoint**: `PUT /api/v1/profile`
**Auth**: User token required

**Request Body**:

```json
{
  "name": "John Updated",
  "phone": "081234567892"
}
```

**Response**:

```json
{
  "success": true,
  "message": "Profile updated successfully",
  "data": {
    "id": 5个子,
    "name": "John Updated",
    "email": "john@example.com",
    "phone": "081234567892"
  }
}
```

---

---

## 3. JUKIR FLOW

### 3.1 Jukir Login

#### Step 1: Jukir Login

**Endpoint**: `POST /api/v1/auth/login`

**Request Body**:

```json
{
  "email": "mal1430",
  "password": "1234"
}
```

**Response**:

```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "refresh_token_string",
    "user": {
      "id": 10,
      "name": "Budi Santoso",
      "email": "mal1430",
      "role": "jukir"
    }
  }
}
```

---

### 3.2 Jukir Dashboard

#### Step 2: Get Jukir Dashboard

**Endpoint**: `GET /api/v1/jukir/dashboard`
**Auth**: Jukir token required

**Response**:

```json
{
  "success": true,
  "message": "Dashboard data retrieved successfully",
  "data": {
    "active_sessions": 5,
    "pending_payments": 3,
    "today_revenue": 150000,
    "total_sessions_today": 25,
    "area": {
      "id": 1,
      "name": "Parkir Mall Palembang"
    }
  }
}
```

---

#### Step 3: Get Pending Payments

**Endpoint**: `GET /api/v1/jukir/pending-payments`
**Auth**: Jukir token required

**Response**:

```json
{
  "success": true,
  "message": "Pending payments retrieved successfully",
  "data": [
    {
      "session_id": 123,
      "vehicle_type": "mobil",
      "plat_nomor": "BG1234XYZ",
      "checkin_time": "2025-01-20T14:30:00Z",
      "duration_minutes": 60,
      "total_cost": 2000
    }
  ]
}
```

---

#### Step 4: Get Active Sessions

**Endpoint**: `GET /api/v1/jukir/active-sessions`
**Auth**: Jukir token required

**Response**:

```json
{
  "success": true,
  "message": "Active sessions retrieved successfully",
  "data": [
    {
      "session_id": 125,
      "vehicle_type": "motor",
      "plat_nomor": "BG5678ABC",
      "checkin_time": "2025-01-20T15:00:00Z",
      "duration_minutes": 30,
      "estimated_cost": 1000
    }
  ]
}
```

---

#### Step 5: Get Vehicle Breakdown

**Endpoint**: `GET /api/v1/jukir/vehicle-breakdown`
**Auth**: Jukir token required

**Response**:

```json
{
  "success": true,
  "message": "Vehicle breakdown retrieved successfully",
  "data": {
    "vehicles_in": 45,
    "vehicles_out": 32,
    "vehicles_by_type": {
      "mobil": { "in": 20, "out": 15 },
      "motor": { "in": 25, "out": 17 }
    }
  }
}
```

---

#### Step 6: Get QR Code

**Endpoint**: `GET /api/v1/jukir/qr-code`
**Auth**: Jukir token required

**Response**:

```json
{
  "success": true,
  "message": "QR code retrieved successfully",
  "data": {
    "qr_token": "QR_MAL1430_20250120143000",
    "area_name": "Parkir Mall Palembang",
    "code": "MAL1430"
  }
}
```

---

#### Step 7: Get Daily Report

**Endpoint**: `GET /api/v1/jukir/daily-report`
**Auth**: Jukir token required

**Response**:

```json
{
  "success": true,
  "message": "Daily report retrieved successfully",
  "data": {
    "date": "2025-01-20",
    "total_sessions": 25,
    "total_revenue": 150000,
    "sessions_by_type": {
      "mobil": 10,
      "motor": 15
    },
    "active_sessions": 5,
    "completed_sessions": 20
  }
}
```

---

### 3.3 Manual Operations

#### Step 8: Manual Check-In

**Endpoint**: `POST /api/v1/jukir/manual-checkin`
**Auth**: Jukir token required

**Request Body**:

```json
{
  "vehicle_type": "mobil",
  "plat_nomor": "BG1234XYZ",
  "latitude": -2.9761,
  "longitude": 104.7754
}
```

**Response**:

```json
{
  "success": true,
  "message": "Manual check-in successful",
  "data": {
    "session_id": 128,
    "checkin_time": "2025-01-20T16:00:00Z",
    "vehicle_type": "mobil",
    "plat_nomor": "BG1234XYZ",
    "area_name": "Parkir Mall Palembang"
  }
}
```

---

#### Step 9: Manual Check-Out

**Endpoint**: `POST /api/v1/jukir/manual-checkout`
**Auth**: Jukir token required

**Request Body**:

```json
{
  "session_id": 128,
  "plat_nomor": "BG1234XYZ",
  "latitude": -2.9761,
  "longitude": tracked
}
```

**Response**:

```json
{
  "success": true,
  "message": "Manual check-out successful",
  "data": {
    "session_id": 128,
    "checkin_time": "2025-01-20T16:00:00Z",
    "checkout_time": "2025-01-20T17:00:00Z",
    "duration_minutes": 60,
    "total_cost": 2000,
    "payment_status": "pending"
  }
}
```

---

#### Step 10: Confirm Payment

**Endpoint**: `POST /api/v1/jukir/confirm-payment`
**Auth**: Jukir token required

**Request Body**:

```json
{
  "session_id": 123,
  "payment_method": "cash"
}
```

**Response**:

```json
{
  "success": true,
  "message": "Payment confirmed successfully",
  "data": {
    "session_id": 123,
    "total_cost": 2000,
    "payment_status": "paid",
    "payment_time": "2025-01-20T15:45:00Z"
  }
}
```

---

---

## 4. API ENDPOINTS SUMMARY

### Authentication Endpoints (No Auth Required)

- `POST /api/v1/auth/register` - Register user/admin/jukir
- `POST /api/v1/auth/login` - Login user/admin/jukir
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout (Auth required)

### Public Parking Endpoints (No Auth Required)

- `GET /api/v1/parking/locations` - Find nearby parking areas
- `POST /api/v1/parking/checkin` - Check-in to parking
- `POST /api/v1/parking/checkout` - Check-out from parking
- `GET /api/v1/parking/active` - Get active session
- `GET /api/v1/parking/history` - Get parking history

### User Profile Endpoints (User Auth Required)

- `GET /api/v1/profile` - Get user profile
- `PUT /api/v1/profile` - Update user profile

### Jukir Endpoints (Jukir Auth Required)

- `GET /api/v1/jukir/dashboard` - Get jukir dashboard
- `GET /api/v1/jukir/pending-payments` - Get pending payments
- `GET /api/v1/jukir/active-sessions` - Get active sessions
- `GET /api/v1/jukir/vehicle-breakdown` - Get vehicle breakdown
- `GET /api/v1/jukir/qr-code` - Get QR code
- `GET /api/v1/jukir/daily-report` - Get daily report
- `POST /api/v1/jukir/confirm-payment` - Confirm payment
- `POST /api/v1/jukir/manual-checkin` - Manual check-in
- `POST /api/v1/jukir/manual-checkout` - Manual check-out
- `GET /api/v1/jukir/events` - SSE events stream

### Admin Endpoints (Admin Auth Required)

**Dashboard & Overview**:

- `GET /api/v1/admin/overview` - Admin overview
- `GET /api/v1/admin/revenue/total` - Total revenue
- `GET /api/v1/admin/chart/data` - Chart data

**Statistics**:

- `GET /api/v1/admin/statistics/vehicles` - Vehicle statistics
- `GET /api/v1/admin/statistics/areas` - Parking area statistics
- `GET /api/v1/admin/statistics/jukirs` - Jukir statistics

**Jukir Management**:

- `GET /api/v1/admin/jukirs` - List jukirs (paginated)
- `GET /api/v1/admin/jukirs/list` - List all jukirs (with filters)
- `GET /api/v1/admin/jukirs/:id` - Get jukir by ID
- `GET /api/v1/admin/jukirs/revenue` - Get jukirs revenue
- `POST /api/v1/admin/jukirs` - Create jukir
- `POST /api/v1/admin/jukirs/manual-revenue` - Add manual revenue
- `PUT /api/v1/admin/jukirs/:id/status` - Update jukir status

**Parking Area Management**:

- `GET /api/v1/admin/areas` - List parking areas
- `GET /api/v1/admin/areas/:id` - Get parking area detail
- `GET /api/v1/admin/areas/:id/transactions` - Get area transactions
- `POST /api/v1/admin/areas` - Create parking area
- `PUT /api/v1/admin/areas/:id` - Update parking area

**Reports & Sessions**:

- `GET /api/v1/admin/reports` - Get reports
- `GET /api/v1/admin/sessions` - Get all sessions
- `GET /api/v1/admin/revenue-table` - Get revenue table
- `GET /api/v1/admin/sse-status` - SSE connection status

---

## 📊 Key Features Summary

### Revenue Tracking

- **Actual Revenue**: Manual input from admin/jukir
- **Estimated Revenue**: Calculated from active sessions
- **Total Revenue**: Sum of actual + estimated

### Filter Capabilities

- **Vehicle Type**: mobil/motor
- **Date Range**: hari_ini/minggu_ini/bulan_ini/tahun_ini
- **Status**: active/inactive/pending (for jukirs and areas)

### Real-Time Features

- SSE (Server-Sent Events) for real-time updates
- Real-time dashboard updates for jukir
- Live active session tracking

### Security Features

- JWT authentication (1 hour expiry)
- Refresh token in Redis (7 days expiry)
- Role-based access control (admin, user, jukir)
- API key protection for all endpoints

---

## 🔐 Authentication Flow

1. User registers or logs in → Gets access_token + refresh_token
2. Use access_token for all authenticated requests
3. When access_token expires → Use refresh_token to get new access_token
4. When refresh_token expires → User must login again
5. Logout invalidates refresh_token in Redis

---

## 📱 Mobile App Integration Points

### User Mobile App:

1. Register/Login
2. Find nearby parking
3. Scan QR code to check-in
4. View active session
5. Scan QR code to check-out
6. View payment invoice
7. View parking history
8. Update profile

### Jukir Mobile App:

1. Login with jukir credentials
2. View dashboard
3. Get QR code to display
4. View pending payments
5. Confirm payments
6. Manual check-in/check-out
7. View daily report
8. Real-time updates via SSE

---

**END OF DOCUMENTATION**
