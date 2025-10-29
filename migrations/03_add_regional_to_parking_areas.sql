-- Migration: Add regional field to parking_areas table
-- This migration adds regional column to support regional grouping of parking areas

-- Add regional column to parking_areas table
ALTER TABLE parking_areas 
ADD COLUMN IF NOT EXISTS regional VARCHAR(50);

-- Create index for regional for better query performance
CREATE INDEX IF NOT EXISTS idx_parking_areas_regional ON parking_areas(regional);

-- Add comment
COMMENT ON COLUMN parking_areas.regional IS 'Regional/regional grouping of the parking area';

