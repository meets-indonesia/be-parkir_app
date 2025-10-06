-- Be-Parkir Database Schema for dbdiagram.io
-- This SQL file can be imported into dbdiagram.io to generate visual ERD

-- Users table - stores all system users (customers, jukirs, admins)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    phone VARCHAR(15) NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'customer',
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Parking areas - locations where parking is available
CREATE TABLE parking_areas (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    address VARCHAR(255) NOT NULL,
    latitude DECIMAL(10, 8) NOT NULL,
    longitude DECIMAL(11, 8) NOT NULL,
    hourly_rate DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL
);

-- Jukirs table - parking attendants with specific areas
CREATE TABLE jukirs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER UNIQUE NOT NULL,
    jukir_code VARCHAR(20) UNIQUE NOT NULL,
    area_id INTEGER NOT NULL,
    qr_token VARCHAR(255) UNIQUE NOT NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (area_id) REFERENCES parking_areas(id) ON DELETE RESTRICT
);

-- Parking sessions - main parking transactions (anonymous)
CREATE TABLE parking_sessions (
    id SERIAL PRIMARY KEY,
    jukir_id INTEGER NULL,
    area_id INTEGER NOT NULL,
    vehicle_type VARCHAR(10) NOT NULL,
    plat_nomor VARCHAR(20) NULL,
    is_manual_record BOOLEAN NOT NULL DEFAULT FALSE,
    checkin_time TIMESTAMP NOT NULL,
    checkout_time TIMESTAMP NULL,
    duration INTEGER NULL,
    total_cost DECIMAL(10, 2) NULL,
    payment_status VARCHAR(20) NOT NULL DEFAULT 'pending',
    session_status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (jukir_id) REFERENCES jukirs(id) ON DELETE SET NULL,
    FOREIGN KEY (area_id) REFERENCES parking_areas(id) ON DELETE RESTRICT
);

-- Note: No user_id in parking_sessions for maximum anonymity
-- plat_nomor is only required for manual records by jukirs
-- QR-based sessions don't require license plate

-- Payments - payment records for parking sessions
CREATE TABLE payments (
    id SERIAL PRIMARY KEY,
    session_id INTEGER NOT NULL,
    amount DECIMAL(10, 2) NOT NULL,
    payment_method VARCHAR(20) NOT NULL,
    confirmed_by INTEGER NULL,
    confirmed_at TIMESTAMP NULL,
    status VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP NULL,
    FOREIGN KEY (session_id) REFERENCES parking_sessions(id) ON DELETE CASCADE,
    FOREIGN KEY (confirmed_by) REFERENCES jukirs(id) ON DELETE SET NULL
);

-- Indexes for better performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_status ON users(status);

CREATE INDEX idx_parking_areas_status ON parking_areas(status);
CREATE INDEX idx_parking_areas_location ON parking_areas(latitude, longitude);

CREATE INDEX idx_jukirs_user_id ON jukirs(user_id);
CREATE INDEX idx_jukirs_jukir_code ON jukirs(jukir_code);
CREATE INDEX idx_jukirs_qr_token ON jukirs(qr_token);
CREATE INDEX idx_jukirs_area_id ON jukirs(area_id);
CREATE INDEX idx_jukirs_status ON jukirs(status);

CREATE INDEX idx_parking_sessions_jukir_id ON parking_sessions(jukir_id);
CREATE INDEX idx_parking_sessions_area_id ON parking_sessions(area_id);
CREATE INDEX idx_parking_sessions_vehicle_type ON parking_sessions(vehicle_type);
CREATE INDEX idx_parking_sessions_plat_nomor ON parking_sessions(plat_nomor);
CREATE INDEX idx_parking_sessions_is_manual_record ON parking_sessions(is_manual_record);
CREATE INDEX idx_parking_sessions_session_status ON parking_sessions(session_status);
CREATE INDEX idx_parking_sessions_payment_status ON parking_sessions(payment_status);
CREATE INDEX idx_parking_sessions_checkin_time ON parking_sessions(checkin_time);

CREATE INDEX idx_payments_session_id ON payments(session_id);
CREATE INDEX idx_payments_confirmed_by ON payments(confirmed_by);
CREATE INDEX idx_payments_status ON payments(status);
CREATE INDEX idx_payments_confirmed_at ON payments(confirmed_at);
