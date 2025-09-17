-- Initial database setup for Parking Digital API
-- This file will be executed when the PostgreSQL container starts

-- Create database if not exists (this will be handled by POSTGRES_DB env var)
-- CREATE DATABASE IF NOT EXISTS parking_app;

-- Create extensions if needed
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- The tables will be created automatically by GORM AutoMigrate
-- This file is here for any manual SQL that might be needed

-- Create indexes for better performance
-- These will be created after the tables are created by GORM

-- Index for users table
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- Index for jukirs table
CREATE INDEX IF NOT EXISTS idx_jukirs_user_id ON jukirs(user_id);
CREATE INDEX IF NOT EXISTS idx_jukirs_jukir_code ON jukirs(jukir_code);
CREATE INDEX IF NOT EXISTS idx_jukirs_qr_token ON jukirs(qr_token);
CREATE INDEX IF NOT EXISTS idx_jukirs_area_id ON jukirs(area_id);
CREATE INDEX IF NOT EXISTS idx_jukirs_status ON jukirs(status);

-- Index for parking_areas table
CREATE INDEX IF NOT EXISTS idx_parking_areas_status ON parking_areas(status);
CREATE INDEX IF NOT EXISTS idx_parking_areas_location ON parking_areas(latitude, longitude);

-- Index for parking_sessions table
CREATE INDEX IF NOT EXISTS idx_parking_sessions_user_id ON parking_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_parking_sessions_jukir_id ON parking_sessions(jukir_id);
CREATE INDEX IF NOT EXISTS idx_parking_sessions_area_id ON parking_sessions(area_id);
CREATE INDEX IF NOT EXISTS idx_parking_sessions_status ON parking_sessions(session_status);
CREATE INDEX IF NOT EXISTS idx_parking_sessions_payment_status ON parking_sessions(payment_status);
CREATE INDEX IF NOT EXISTS idx_parking_sessions_checkin_time ON parking_sessions(checkin_time);

-- Index for payments table
CREATE INDEX IF NOT EXISTS idx_payments_session_id ON payments(session_id);
CREATE INDEX IF NOT EXISTS idx_payments_confirmed_by ON payments(confirmed_by);
CREATE INDEX IF NOT EXISTS idx_payments_status ON payments(status);
CREATE INDEX IF NOT EXISTS idx_payments_confirmed_at ON payments(confirmed_at);
