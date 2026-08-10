-- Add user_id to expenses, backfill from the owning resident_location, enforce NOT NULL.
-- Idempotent: safe to re-run.

-- Step 1: add column as nullable (existing rows have no user_id)
ALTER TABLE expenses
  ADD COLUMN IF NOT EXISTS user_id TEXT REFERENCES "user"(id) ON DELETE CASCADE;

-- Step 2: backfill from the owning resident_location
UPDATE expenses e
SET user_id = rl.user_id
FROM resident_locations rl
WHERE rl.id = e.resident_location_id
  AND e.user_id IS NULL;

-- Step 3: fail loudly if any expense could not be backfilled
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM expenses WHERE user_id IS NULL) THEN
        RAISE EXCEPTION 'expenses.user_id backfill failed: % expense(s) have no resident_location or owner',
            (SELECT COUNT(*) FROM expenses WHERE user_id IS NULL);
    END IF;
END $$;

-- Step 4: enforce NOT NULL now that all rows have a value
ALTER TABLE expenses ALTER COLUMN user_id SET NOT NULL;

-- Step 5: index for ownership lookups
CREATE INDEX IF NOT EXISTS expenses_user_id_idx ON expenses (user_id);