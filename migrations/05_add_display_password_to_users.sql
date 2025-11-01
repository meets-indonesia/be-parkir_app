-- Migration: Add display_password column to users table
-- This column stores plain text password for jukir users to display in admin panel

-- Add display_password column
ALTER TABLE users 
ADD COLUMN IF NOT EXISTS display_password VARCHAR(20);

-- Add comment
COMMENT ON COLUMN users.display_password IS 'Plain text password for display purposes (jukir only)';

