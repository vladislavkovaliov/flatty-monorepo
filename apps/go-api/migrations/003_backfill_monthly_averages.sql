-- Backfill expense_monthly_averages from expenses data
-- Idempotent: safe to re-run

ALTER TABLE expense_monthly_averages
  ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP;

INSERT INTO expense_monthly_averages (month, year, average_amount, expense_count, updated_at, created_at)
SELECT
    month,
    year,
    SUM(amount) / COUNT(*) AS average_amount,
    COUNT(*) AS expense_count,
    CURRENT_TIMESTAMP AS updated_at,
    MIN(created_at) AS created_at
FROM expenses
GROUP BY month, year
ON CONFLICT (month, year) DO UPDATE SET
    average_amount = EXCLUDED.average_amount,
    expense_count  = EXCLUDED.expense_count,
    updated_at     = CURRENT_TIMESTAMP,
    created_at     = LEAST(expense_monthly_averages.created_at, EXCLUDED.created_at);
