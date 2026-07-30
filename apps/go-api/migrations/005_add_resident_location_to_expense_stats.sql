-- Scope expense_monthly_totals / expense_monthly_averages by resident_location_id.
-- Previously both tables were keyed only by (month, year), aggregating expenses
-- across every resident_location (and therefore every owning user) into one row.
-- Idempotent: safe to re-run.

ALTER TABLE expense_monthly_totals
  ADD COLUMN IF NOT EXISTS resident_location_id BIGINT REFERENCES resident_locations(id) ON DELETE CASCADE;

ALTER TABLE expense_monthly_averages
  ADD COLUMN IF NOT EXISTS resident_location_id BIGINT REFERENCES resident_locations(id) ON DELETE CASCADE;

ALTER TABLE expense_monthly_totals
  DROP CONSTRAINT IF EXISTS expense_monthly_totals_pkey;

ALTER TABLE expense_monthly_averages
  DROP CONSTRAINT IF EXISTS expense_monthly_averages_pkey;

-- Existing rows were global aggregates and cannot be attributed to a single
-- resident_location; rebuild both tables from the expenses source of truth.
TRUNCATE TABLE expense_monthly_totals;
TRUNCATE TABLE expense_monthly_averages;

INSERT INTO expense_monthly_totals (resident_location_id, month, year, total_spent, updated_at)
SELECT
    resident_location_id,
    month,
    year,
    SUM(amount) AS total_spent,
    CURRENT_TIMESTAMP AS updated_at
FROM expenses
GROUP BY resident_location_id, month, year;

INSERT INTO expense_monthly_averages (resident_location_id, month, year, average_amount, expense_count, updated_at)
SELECT
    resident_location_id,
    month,
    year,
    SUM(amount) / COUNT(*) AS average_amount,
    COUNT(*) AS expense_count,
    CURRENT_TIMESTAMP AS updated_at
FROM expenses
GROUP BY resident_location_id, month, year;

ALTER TABLE expense_monthly_totals
  ALTER COLUMN resident_location_id SET NOT NULL;

ALTER TABLE expense_monthly_averages
  ALTER COLUMN resident_location_id SET NOT NULL;

ALTER TABLE expense_monthly_totals
  ADD PRIMARY KEY (resident_location_id, month, year);

ALTER TABLE expense_monthly_averages
  ADD PRIMARY KEY (resident_location_id, month, year);
