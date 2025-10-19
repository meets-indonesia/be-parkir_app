# Mobile App Implementation Guide

Complete guide for implementing the Be-Parkir mobile applications.

---

## 📱 **Two Mobile Apps Required**

1. **Customer App** (Anonymous Parking)
2. **Jukir App** (Parking Attendant)

---

## 🎯 **1. Customer/Anonymous Parking App**

### **App Features**

- Find nearby parking areas
- Scan QR code to check-in
- View active parking session
- Scan QR code to check-out
- View parking history (optional)

### **Endpoints to Implement**

| Endpoint                                                            | Method | Purpose          | Screen         |
| ------------------------------------------------------------------- | ------ | ---------------- | -------------- |
| `GET /parking/locations?latitude={lat}&longitude={lon}&radius={km}` | GET    | Find parking     | Map Screen     |
| `POST /parking/checkin`                                             | POST   | Check-in via QR  | Scanner Screen |
| `GET /parking/active?qr_token={token}`                              | GET    | View session     | Active Screen  |
| `POST /parking/checkout`                                            | POST   | Check-out via QR | Scanner Screen |
| `GET /parking/history?plat_nomor={plate}&limit=10&offset=0`         | GET    | View history     | History Screen |

**Headers Required:**

```javascript
{
  'X-API-Key': 'be-parkir-api-key-2025',
  'Content-Type': 'application/json'
}
```

**Note:** ✅ No authentication required! (Anonymous parking)

---

### **Customer App - Complete Example**

```javascript
// services/parkingApi.js
const API_BASE = "http://localhost:8080/api/v1";
const API_KEY = "be-parkir-api-key-2025";

export const ParkingAPI = {
  // Find nearby parking areas
  findNearbyParking: async (latitude, longitude, radius = 1.0) => {
    const response = await fetch(
      `${API_BASE}/parking/locations?latitude=${latitude}&longitude=${longitude}&radius=${radius}`,
      {
        headers: { "X-API-Key": API_KEY },
      }
    );
    return response.json();
  },

  // Check-in to parking
  checkin: async (qrToken, latitude, longitude, vehicleType) => {
    const response = await fetch(`${API_BASE}/parking/checkin`, {
      method: "POST",
      headers: {
        "X-API-Key": API_KEY,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        qr_token: qrToken,
        latitude,
        longitude,
        vehicle_type: vehicleType, // 'mobil' or 'motor'
      }),
    });
    return response.json();
  },

  // Get active session
  getActiveSession: async (qrToken) => {
    const response = await fetch(
      `${API_BASE}/parking/active?qr_token=${qrToken}`,
      {
        headers: { "X-API-Key": API_KEY },
      }
    );
    return response.json();
  },

  // Check-out from parking
  checkout: async (qrToken, latitude, longitude) => {
    const response = await fetch(`${API_BASE}/parking/checkout`, {
      method: "POST",
      headers: {
        "X-API-Key": API_KEY,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        qr_token: qrToken,
        latitude,
        longitude,
      }),
    });
    return response.json();
  },

  // Get parking history
  getHistory: async (platNomor, limit = 10, offset = 0) => {
    const response = await fetch(
      `${API_BASE}/parking/history?plat_nomor=${platNomor}&limit=${limit}&offset=${offset}`,
      {
        headers: { "X-API-Key": API_KEY },
      }
    );
    return response.json();
  },
};
```

---

## 🅿️ **2. Jukir (Parking Attendant) App**

### **App Features**

- Login with email/password
- View dashboard with statistics
- Show QR code to customers
- View active parking sessions (real-time)
- Manual check-in (for customers without phones)
- Manual check-out
- View pending payments
- Confirm payments (cash/QRIS)
- Daily revenue reports
- **Real-time notifications via SSE** ⚡

---

### **Endpoints to Implement**

#### **A. Authentication (1 endpoint)**

| Endpoint           | Method | When to Call |
| ------------------ | ------ | ------------ |
| `POST /auth/login` | POST   | Login screen |

---

#### **B. Dashboard & Info (5 endpoints)**

| Endpoint                                  | Method | When to Call          |
| ----------------------------------------- | ------ | --------------------- |
| `GET /jukir/dashboard`                    | GET    | Dashboard screen load |
| `GET /jukir/qr-code`                      | GET    | QR code screen load   |
| `GET /jukir/active-sessions`              | GET    | Sessions screen load  |
| `GET /jukir/pending-payments`             | GET    | Payments screen load  |
| `GET /jukir/daily-report?date=YYYY-MM-DD` | GET    | Reports screen load   |

---

#### **C. Operations (3 endpoints)**

| Endpoint                      | Method | When to Call           |
| ----------------------------- | ------ | ---------------------- |
| `POST /jukir/manual-checkin`  | POST   | Manual entry button    |
| `POST /jukir/manual-checkout` | POST   | Manual exit button     |
| `POST /jukir/confirm-payment` | POST   | Payment confirm button |

---

#### **D. Real-Time Updates (1 SSE endpoint)** ⚡

| Endpoint            | Method    | When to Call          |
| ------------------- | --------- | --------------------- |
| `GET /jukir/events` | GET (SSE) | **ONCE on app start** |

**This is the magic endpoint!** ✨

- Call **ONCE** when app starts
- Connection stays **OPEN**
- Server **pushes** events when they happen
- **Event-driven** notifications

---

#### **E. Profile (2 endpoints)**

| Endpoint       | Method | When to Call        |
| -------------- | ------ | ------------------- |
| `GET /profile` | GET    | Profile screen load |
| `PUT /profile` | PUT    | Save profile button |

---

### **Jukir App - Complete Implementation**

```javascript
// services/jukirApi.js
import AsyncStorage from "@react-native-async-storage/async-storage";

const API_BASE = "http://localhost:8080/api/v1";
const API_KEY = "be-parkir-api-key-2025";

const getToken = async () => {
  return await AsyncStorage.getItem("jukir_token");
};

export const JukirAPI = {
  // ==========================================
  // Authentication
  // ==========================================

  login: async (email, password) => {
    const response = await fetch(`${API_BASE}/auth/login`, {
      method: "POST",
      headers: {
        "X-API-Key": API_KEY,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ email, password }),
    });
    const data = await response.json();

    if (data.success) {
      await AsyncStorage.setItem("jukir_token", data.data.access_token);
      await AsyncStorage.setItem("refresh_token", data.data.refresh_token);
    }

    return data;
  },

  logout: async () => {
    const token = await getToken();
    const response = await fetch(`${API_BASE}/auth/logout`, {
      method: "POST",
      headers: {
        "X-API-Key": API_KEY,
        Authorization: `Bearer ${token}`,
      },
    });

    await AsyncStorage.removeItem("jukir_token");
    await AsyncStorage.removeItem("refresh_token");

    return response.json();
  },

  // ==========================================
  // Dashboard & Info
  // ==========================================

  getDashboard: async () => {
    const token = await getToken();
    const response = await fetch(`${API_BASE}/jukir/dashboard`, {
      headers: {
        "X-API-Key": API_KEY,
        Authorization: `Bearer ${token}`,
      },
    });
    return response.json();
  },

  getQRCode: async () => {
    const token = await getToken();
    const response = await fetch(`${API_BASE}/jukir/qr-code`, {
      headers: {
        "X-API-Key": API_KEY,
        Authorization: `Bearer ${token}`,
      },
    });
    return response.json();
  },

  getActiveSessions: async () => {
    const token = await getToken();
    const response = await fetch(`${API_BASE}/jukir/active-sessions`, {
      headers: {
        "X-API-Key": API_KEY,
        Authorization: `Bearer ${token}`,
      },
    });
    return response.json();
  },

  getPendingPayments: async () => {
    const token = await getToken();
    const response = await fetch(`${API_BASE}/jukir/pending-payments`, {
      headers: {
        "X-API-Key": API_KEY,
        Authorization: `Bearer ${token}`,
      },
    });
    return response.json();
  },

  getDailyReport: async (date) => {
    const token = await getToken();
    const response = await fetch(
      `${API_BASE}/jukir/daily-report?date=${date}`,
      {
        headers: {
          "X-API-Key": API_KEY,
          Authorization: `Bearer ${token}`,
        },
      }
    );
    return response.json();
  },

  // ==========================================
  // Manual Operations
  // ==========================================

  manualCheckin: async (platNomor, vehicleType, waktuMasuk) => {
    const token = await getToken();
    const response = await fetch(`${API_BASE}/jukir/manual-checkin`, {
      method: "POST",
      headers: {
        "X-API-Key": API_KEY,
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        plat_nomor: platNomor,
        vehicle_type: vehicleType, // 'mobil' or 'motor'
        waktu_masuk: waktuMasuk || new Date().toISOString(),
      }),
    });
    return response.json();
  },

  manualCheckout: async (sessionId, waktuKeluar) => {
    const token = await getToken();
    const response = await fetch(`${API_BASE}/jukir/manual-checkout`, {
      method: "POST",
      headers: {
        "X-API-Key": API_KEY,
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        session_id: sessionId,
        waktu_keluar: waktuKeluar || new Date().toISOString(),
      }),
    });
    return response.json();
  },

  confirmPayment: async (sessionId, paymentMethod) => {
    const token = await getToken();
    const response = await fetch(`${API_BASE}/jukir/confirm-payment`, {
      method: "POST",
      headers: {
        "X-API-Key": API_KEY,
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        session_id: sessionId,
        payment_method: paymentMethod, // 'cash' or 'qris'
      }),
    });
    return response.json();
  },

  // ==========================================
  // Profile
  // ==========================================

  getProfile: async () => {
    const token = await getToken();
    const response = await fetch(`${API_BASE}/profile`, {
      headers: {
        "X-API-Key": API_KEY,
        Authorization: `Bearer ${token}`,
      },
    });
    return response.json();
  },

  updateProfile: async (name, phone) => {
    const token = await getToken();
    const response = await fetch(`${API_BASE}/profile`, {
      method: "PUT",
      headers: {
        "X-API-Key": API_KEY,
        Authorization: `Bearer ${token}`,
        "Content-Type": "application/json",
      },
      body: JSON.stringify({ name, phone }),
    });
    return response.json();
  },
};
```

---

## ⚡ **SSE Implementation (Event-Driven)**

### **The Special Endpoint**

**Endpoint:** `GET /api/v1/jukir/events`

**How to Use:**

```javascript
// hooks/useJukirEventStream.js
import { useState, useEffect, useRef } from "react";
import EventSource from "react-native-event-source"; // Install this!

const useJukirEventStream = (jukirToken, apiKey) => {
  const [connected, setConnected] = useState(false);
  const [events, setEvents] = useState([]);
  const eventSourceRef = useRef(null);

  useEffect(() => {
    if (!jukirToken || !apiKey) return;

    // ⚡ CONNECT ONCE when app starts
    const url = "http://localhost:8080/api/v1/jukir/events";

    eventSourceRef.current = new EventSource(url, {
      headers: {
        Authorization: `Bearer ${jukirToken}`,
        "X-API-Key": apiKey,
      },
    });

    // Connection opened
    eventSourceRef.current.addEventListener("open", () => {
      console.log("✅ SSE Connected");
      setConnected(true);
    });

    // Connection established
    eventSourceRef.current.addEventListener("connected", (e) => {
      const data = JSON.parse(e.data);
      console.log("Connected:", data);
    });

    // ⚡ RECEIVE EVENTS (EVENT-DRIVEN!)
    eventSourceRef.current.addEventListener("message", (e) => {
      try {
        const eventData = JSON.parse(e.data);

        // Add to events list
        setEvents((prev) => [eventData, ...prev]);

        // Handle event based on type
        switch (eventData.type) {
          case "session_update":
            handleSessionUpdate(eventData.data);
            break;
          case "payment_confirmed":
            handlePaymentConfirmed(eventData.data);
            break;
        }
      } catch (error) {
        console.error("Error parsing event:", error);
      }
    });

    // Handle ping (keep-alive)
    eventSourceRef.current.addEventListener("ping", (e) => {
      console.log("Ping received");
    });

    // Handle errors
    eventSourceRef.current.addEventListener("error", (e) => {
      console.error("SSE Error:", e);
      setConnected(false);

      // Will auto-reconnect
    });

    // Cleanup on unmount
    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
    };
  }, [jukirToken, apiKey]);

  const handleSessionUpdate = (data) => {
    // Show notification
    showNotification(
      "Session Updated",
      `${data.plat_nomor || "Vehicle"} checked out - IDR ${data.total_cost}`
    );
  };

  const handlePaymentConfirmed = (data) => {
    showNotification(
      "Payment Received",
      `IDR ${data.amount} via ${data.payment_method}`
    );
  };

  return { connected, events };
};

export default useJukirEventStream;
```

---

## 📦 **Required npm Packages**

### **For React Native**

```bash
# Core dependencies
npm install @react-navigation/native
npm install @react-navigation/stack
npm install @react-native-async-storage/async-storage
npm install react-native-qrcode-scanner
npm install react-native-camera
npm install react-native-qrcode-svg

# For SSE (Event-Driven Updates) ⚡
npm install react-native-event-source

# For notifications
npm install react-native-push-notification
npm install @react-native-community/push-notification-ios

# For date/time
npm install date-fns

# For forms
npm install react-hook-form
```

---

## 🔑 **Complete Endpoint Reference**

### **Jukir App - All 12 Endpoints**

```javascript
const JUKIR_ENDPOINTS = {
  // 1. Authentication
  LOGIN: {
    url: "/auth/login",
    method: "POST",
    body: { email, password },
    auth: false, // No token needed
    callWhen: "Login button pressed",
  },

  // 2. Dashboard
  DASHBOARD: {
    url: "/jukir/dashboard",
    method: "GET",
    auth: true, // Token required
    callWhen: "Dashboard screen loads",
  },

  // 3. QR Code
  QR_CODE: {
    url: "/jukir/qr-code",
    method: "GET",
    auth: true,
    callWhen: "QR screen loads",
  },

  // 4. Active Sessions
  ACTIVE_SESSIONS: {
    url: "/jukir/active-sessions",
    method: "GET",
    auth: true,
    callWhen: "Sessions screen loads",
  },

  // 5. Pending Payments
  PENDING_PAYMENTS: {
    url: "/jukir/pending-payments",
    method: "GET",
    auth: true,
    callWhen: "Payments screen loads",
  },

  // 6. Daily Report
  DAILY_REPORT: {
    url: "/jukir/daily-report?date=YYYY-MM-DD",
    method: "GET",
    auth: true,
    callWhen: "Reports screen loads",
  },

  // 7. Manual Check-in
  MANUAL_CHECKIN: {
    url: "/jukir/manual-checkin",
    method: "POST",
    body: { plat_nomor, vehicle_type, waktu_masuk },
    auth: true,
    callWhen: "Manual check-in button pressed",
  },

  // 8. Manual Check-out
  MANUAL_CHECKOUT: {
    url: "/jukir/manual-checkout",
    method: "POST",
    body: { session_id, waktu_keluar },
    auth: true,
    callWhen: "Manual check-out button pressed",
  },

  // 9. Confirm Payment
  CONFIRM_PAYMENT: {
    url: "/jukir/confirm-payment",
    method: "POST",
    body: { session_id, payment_method },
    auth: true,
    callWhen: "Confirm payment button pressed",
  },

  // 10. Get Profile
  GET_PROFILE: {
    url: "/profile",
    method: "GET",
    auth: true,
    callWhen: "Profile screen loads",
  },

  // 11. Update Profile
  UPDATE_PROFILE: {
    url: "/profile",
    method: "PUT",
    body: { name, phone },
    auth: true,
    callWhen: "Save profile button pressed",
  },

  // 12. SSE Events ⚡ (MOST IMPORTANT!)
  EVENTS: {
    url: "/jukir/events",
    method: "GET (SSE)",
    auth: true,
    callWhen: "ONCE on app start",
    staysConnected: true, // Connection stays open!
    eventDriven: true, // Pushes events when they happen!
  },
};
```

---

## 🎯 **Implementation Checklist**

### **Phase 1: Basic Jukir App**

- [ ] Login screen
- [ ] Dashboard screen
- [ ] QR code display screen
- [ ] Pending payments list
- [ ] Payment confirmation

**Endpoints needed:** 5

1. POST /auth/login
2. GET /jukir/dashboard
3. GET /jukir/qr-code
4. GET /jukir/pending-payments
5. POST /jukir/confirm-payment

---

### **Phase 2: Real-Time Updates** ⚡

- [ ] Install `react-native-event-source`
- [ ] Implement SSE hook
- [ ] Connect on app start
- [ ] Handle events
- [ ] Show notifications

**Endpoints needed:** 1 6. GET /jukir/events (SSE)

---

### **Phase 3: Manual Operations**

- [ ] Manual check-in form
- [ ] Manual check-out form
- [ ] Active sessions list

**Endpoints needed:** 3 7. POST /jukir/manual-checkin 8. POST /jukir/manual-checkout 9. GET /jukir/active-sessions

---

### **Phase 4: Reports & Profile**

- [ ] Daily reports screen
- [ ] Profile management

**Endpoints needed:** 3 10. GET /jukir/daily-report 11. GET /profile 12. PUT /profile

---

## 📊 **Quick Reference Card**

### **For Mobile Developer**

```
Total Endpoints: 12

Traditional REST: 11
  - Call when needed
  - Request → Response
  - Normal HTTP

SSE (Special): 1 ⚡
  - Call ONCE on app start
  - Connection stays OPEN
  - Server pushes events
  - EVENT-DRIVEN!

Most Important:
  Priority 1: Login, Dashboard, QR Code, Payments (5 endpoints)
  Priority 2: SSE Events (1 endpoint) ⚡
  Priority 3: Manual operations (3 endpoints)
  Priority 4: Reports & Profile (3 endpoints)
```

---

## 🔗 **Swagger Documentation**

**Interactive API Documentation:**  
🌐 http://localhost:8080/swagger/index.html

This provides:

- ✅ All endpoint details
- ✅ Request/response examples
- ✅ Try it out functionality
- ✅ Schema definitions
- ✅ Authentication info

---

## 📱 **Mobile App Structure**

```
JukirApp/
├── src/
│   ├── services/
│   │   ├── api.js              ← All 11 REST endpoints
│   │   └── eventStream.js      ← SSE connection
│   ├── hooks/
│   │   └── useJukirEvents.js   ← SSE hook (call once!)
│   ├── screens/
│   │   ├── LoginScreen.js
│   │   ├── DashboardScreen.js  ← Uses SSE hook
│   │   ├── QRCodeScreen.js
│   │   ├── ActiveSessionsScreen.js
│   │   ├── ManualCheckinScreen.js
│   │   ├── ManualCheckoutScreen.js
│   │   ├── PendingPaymentsScreen.js
│   │   ├── DailyReportScreen.js
│   │   └── ProfileScreen.js
│   └── components/
│       ├── SessionCard.js
│       ├── PaymentCard.js
│       └── QRCodeDisplay.js
└── App.js                      ← Connect SSE on app start
```

---

## 🎯 **Final Summary**

**What to implement in mobile:**

### **Traditional Endpoints (11):**

Call these when user opens a screen or presses a button.

### **SSE Endpoint (1):** ⚡

Call **ONCE** when app starts, then:

- ✅ Connection stays open
- ✅ Server pushes events
- ✅ App receives updates instantly
- ✅ Event-driven notifications

---

**Next Steps:**

1. Open Swagger: http://localhost:8080/swagger/index.html
2. Explore all endpoints interactively
3. Implement in your mobile app
4. Use SSE for real-time updates!

---

**Status**: ✅ **All Documentation Complete**  
**Swagger**: ✅ **Available at /swagger/index.html**  
**Mobile Guide**: ✅ **Complete**  
**SSE Examples**: ✅ **Ready to use**
