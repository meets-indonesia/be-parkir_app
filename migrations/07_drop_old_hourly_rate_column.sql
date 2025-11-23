-- Drop old hourly_rate column
-- Migration: 07_drop_old_hourly_rate_column.sql
-- This migration removes the old hourly_rate column after ensuring all data has been migrated

-- First, ensure all existing data has been migrated to the new columns
-- (in case migration 06 was run before this)
UPDATE parking_areas 
SET 
    hourly_rate_mobil = COALESCE(hourly_rate_mobil, COALESCE(hourly_rate, 0.00)),
    hourly_rate_motor = COALESCE(hourly_rate_motor, COALESCE(hourly_rate, 0.00))
WHERE (hourly_rate_mobil = 0.00 OR hourly_rate_motor = 0.00) 
  AND hourly_rate IS NOT NULL;

-- Drop the old hourly_rate column
ALTER TABLE parking_areas DROP COLUMN IF EXISTS hourly_rate;

