-- Add separate hourly rates for mobil and motor
-- Migration: 06_add_separate_rates_for_vehicle_types.sql

-- Add new columns for vehicle-specific rates
ALTER TABLE parking_areas 
ADD COLUMN IF NOT EXISTS hourly_rate_mobil DECIMAL(10, 2) NOT NULL DEFAULT 0.00,
ADD COLUMN IF NOT EXISTS hourly_rate_motor DECIMAL(10, 2) NOT NULL DEFAULT 0.00;

-- Migrate existing data: copy hourly_rate to both new columns
UPDATE parking_areas 
SET 
    hourly_rate_mobil = hourly_rate,
    hourly_rate_motor = hourly_rate
WHERE hourly_rate_mobil = 0.00 OR hourly_rate_motor = 0.00;

-- Add comments
COMMENT ON COLUMN parking_areas.hourly_rate_mobil IS 'Flat rate for mobil (car) parking';
COMMENT ON COLUMN parking_areas.hourly_rate_motor IS 'Flat rate for motor (motorcycle) parking';

-- Remove old hourly_rate column (after migration is complete)
-- Note: Uncomment the following line after verifying the migration works correctly
-- ALTER TABLE parking_areas DROP COLUMN IF EXISTS hourly_rate;

