from app.config import load_config
from app.db.connection import create_pool
from app.db.expenses import list_expenses, get_monthly_totals, get_month_avarages, get_year_months

cfg = load_config()
pool = create_pool(cfg.database_url)
print(list_expenses(pool, resident_location_id=1))
print(get_monthly_totals(pool, resident_location_id=1))
print(get_month_avarages(pool, resident_location_id=1))
print(get_year_months(pool, resident_location_id=1))
# print(get_monthly_totals(pool, 1, 2023))
pool.close()