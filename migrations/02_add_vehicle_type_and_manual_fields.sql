-- Migration: Add vehicle_type and manual record fields to parking_sessions table
-- This migration adds support for vehicle types (mobil/motor) and manual record functionality

-- Add vehicle_type column
ALTER TABLE parking_sessions 
ADD COLUMN vehicle_type VARCHAR(10) NOT NULL DEFAULT 'mobil';

-- Add plat_nomor column
ALTER TABLE parking_sessions 
ADD COLUMN plat_nomor VARCHAR(20) NOT NULL DEFAULT '';

-- Add is_manual_record column
ALTER TABLE parking_sessions 
ADD COLUMN is_manual_record BOOLEAN NOT NULL DEFAULT FALSE;

-- Update existing records to have default values
UPDATE parking_sessions 
SET vehicle_type = 'mobil', 
    plat_nomor = 'UNKNOWN', 
    is_manual_record = FALSE 
WHERE vehicle_type = 'mobil' AND plat_nomor = '';

-- Add check constraint for vehicle_type
ALTER TABLE parking_sessions 
ADD CONSTRAINT chk_vehicle_type 
CHECK (vehicle_type IN ('mobil', 'motor'));

-- Create index for vehicle_type for better query performance
CREATE INDEX IF NOT EXISTS idx_parking_sessions_vehicle_type ON parking_sessions(vehicle_type);

-- Create index for plat_nomor for better query performance
CREATE INDEX IF NOT EXISTS idx_parking_sessions_plat_nomor ON parking_sessions(plat_nomor);

-- Create index for is_manual_record for better query performance
CREATE INDEX IF NOT EXISTS idx_parking_sessions_is_manual_record ON parking_sessions(is_manual_record);

-- Remove timeout status from session_status constraint (if it exists)
-- Note: This might need to be adjusted based on your current database schema
-- ALTER TABLE parking_sessions DROP CONSTRAINT IF EXISTS chk_session_status;
-- ALTER TABLE parking_sessions ADD CONSTRAINT chk_session_status CHECK (session_status IN ('active', 'pending_payment', 'completed', 'cancelled'));

-- Update session_status constraint to remove timeout
-- This is a more complex operation that might require recreating the table
-- For now, we'll leave the existing constraint and handle timeout removal in application logic

COMMENT ON COLUMN parking_sessions.vehicle_type IS 'Type of vehicle: mobil or motor';
COMMENT ON COLUMN parking_sessions.plat_nomor IS 'License plate number of the vehicle';
COMMENT ON COLUMN parking_sessions.is_manual_record IS 'Whether this session was created manually by jukir';
