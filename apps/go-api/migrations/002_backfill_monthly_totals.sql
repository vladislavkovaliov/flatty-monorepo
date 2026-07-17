-- Backfill expense_monthly_totals from expenses data
-- Idempotent: safe to re-run

ALTER TABLE expense_monthly_totals
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

INSERT INTO expense_monthly_totals (month, year, total_spent, updated_at, created_at)
SELECT
    month,
    year,
    SUM(amount) AS total_spent,
    CURRENT_TIMESTAMP AS updated_at,
    MIN(created_at) AS created_at
FROM expenses
GROUP BY month, year
ON CONFLICT (month, year) DO UPDATE SET
    total_spent = EXCLUDED.total_spent,
    updated_at  = CURRENT_TIMESTAMP,
    created_at  = LEAST(expense_monthly_totals.created_at, EXCLUDED.created_at);
