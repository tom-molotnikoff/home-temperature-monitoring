-- Add config bag column and migrate URL data

-- 1. Add config column (JSON TEXT) with default empty object
ALTER TABLE sensors ADD COLUMN config TEXT NOT NULL DEFAULT '{}';

-- 2. Migrate existing URL data into config bag
UPDATE sensors SET config = json_object('url', url) WHERE url IS NOT NULL AND url != '';

-- 3. Drop the old url column
ALTER TABLE sensors DROP COLUMN url;
