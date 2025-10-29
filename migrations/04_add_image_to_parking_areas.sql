-- +migrate Up
ALTER TABLE parking_areas
ADD COLUMN IF NOT EXISTS image TEXT NULL;

-- +migrate Down
ALTER TABLE parking_areas
DROP COLUMN IF EXISTS image;


