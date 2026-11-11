from decimal import Decimal

from psycopg.rows import dict_row

from app.db.connection import Pool
from app.db.types import (
    CategoryRow,
    ExpenseRow,
    MonthlyAverageRow,
    MonthlyTotalRow,
    YearMonthRow,
)


def _money(value: Decimal | float | int | None) -> float:
    """ NUMERIC/Decimal -> float. None -> 0.0"""
    if value is not None:
        return float(value)
    else:
        return 0.0

def get_monthly_totals(
    pool: Pool,
    resident_location_id: int,
    year: int | None = None,
    month: int | None = None,
) -> list[MonthlyTotalRow]:
    """Return monthly spending totals for a location.

    Args:
      resident_location_id: location (apartment) id.
      year: year to filter by, if any.
      month: month 1-12 to filter by, if any.
    """
    sql = """
        select month, year, total_spent
        from expense_monthly_totals
        where resident_location_id = %s
            and (%s::int IS NULL OR year = %s)
            and (%s::int IS NULL OR month = %s)
        order BY year DESC, month DESC
    """

    with pool.connection() as conn, conn.cursor(row_factory=dict_row) as cur:
        cur.execute(sql, (resident_location_id, year, year, month, month))
        
        rows = cur.fetchall()

    return [
        {
            "month": int(r["month"]),
            "year": int(r["year"]),
            "total_spent": _money(r["total_spent"]),
        }
        for r in rows
    ]

def list_categories(
    pool: Pool,
) -> list[CategoryRow]:
    """Return the list of expense categories.

    Returns:
      List of categories with id, name, description fields.
    """
    sql = """
        SELECT id, name, description FROM categories ORDER BY name
    """

    with pool.connection() as conn, conn.cursor(row_factory=dict_row) as cur:
        cur.execute(sql)
        rows = cur.fetchall()

    return [
        {
            "id": int(r["id"]),
            "name": str(r["name"]),
            "description": str(r["description"]),
        }
        for r in rows
    ]

def list_expenses(
    pool: Pool,
    resident_location_id: int,
    year: int | None = None,
    month: int | None = None,
    category_id: int | None = None,
    limit: int = 50,
    offset: int = 0,
) -> list[ExpenseRow]:
    """Return a location's expenses with filters and pagination.

    Args:
      resident_location_id: location (apartment) id.
      year: year to filter by, if any.
      month: month 1-12 to filter by, if any.
      category_id: category id to filter by, if any.
      limit: max records (1-100, default 50).
      offset: pagination offset (default 0).

    Returns:
      List of expenses with id, category_id, category_name, amount,
      description, month, year fields.
    """
    where = ["e.resident_location_id = %s"]
    params: list[object] = [resident_location_id]

    if year is not None:
        where.append("e.year = %s")
        params.append(year)

    if month is not None:
        where.append("e.month = %s")
        params.append(month)
    
    if category_id is not None:
        where.append("e.category_id = %s")
        params.append(category_id)

    limit = max(1, min(limit, 100))   # guard against limit=100000 from the LLM
    offset = max(0, offset)

    sql = f"""
        SELECT e.id, e.category_id, c.name AS category_name,
               e.amount, e.description, e.month, e.year
        FROM expenses e
        JOIN categories c ON c.id = e.category_id
        WHERE {' AND '.join(where)}
        ORDER BY e.year DESC, e.month DESC, e.id DESC
        LIMIT %s OFFSET %s
    """
    
    params.extend([limit, offset])

    with pool.connection() as conn, conn.cursor(row_factory=dict_row) as cur:
        cur.execute(sql, params)
        rows = cur.fetchall()

    return [
        {
            "id": int(r["id"]),
            "category_id": int(r["category_id"]),
            "category_name": r["category_name"],
            "amount": _money(r["amount"]),
            "description": r["description"],
            "month": int(r["month"]),
            "year": int(r["year"]),
        }
        for r in rows
    ]

def get_month_averages(
    pool: Pool,
    resident_location_id: int
) -> list[MonthlyAverageRow]:
    """Return a location's monthly average spending.

    Args:
      resident_location_id: location (apartment) id.

    Returns:
      List of records with month, year, average_amount, expense_count fields.
    """
    sql = """
        SELECT month, year, average_amount, expense_count
        FROM expense_monthly_averages
        WHERE resident_location_id = %s
        ORDER BY year DESC, month DESC
    """

    with pool.connection() as conn, conn.cursor(row_factory=dict_row) as cur:
        cur.execute(sql, (
                resident_location_id,
            ),
        )
        rows = cur.fetchall()

    return [
        {
            "month": int(r["month"]),
            "year": int(r["year"]),
            "average_amount": _money(r["average_amount"]),
            "expense_count": int(r["expense_count"]),
        }
        for r in rows
    ]

def get_year_months(
    pool: Pool,
    resident_location_id: int
) -> list[YearMonthRow]:
    """Return the (year, month) pairs for which a location has expenses.

    Args:
      resident_location_id: location (apartment) id.

    Returns:
      List of records with year, month fields.
    """
    sql = """
        SELECT DISTINCT year, month
        FROM expenses
        WHERE resident_location_id = %s
        ORDER BY year DESC, month DESC
    """

    with pool.connection() as conn, conn.cursor(row_factory=dict_row) as cur:
        cur.execute(
            sql,
            (resident_location_id,),
        )
        rows = cur.fetchall()

    return [
        {
            "year": int(r["year"]),
            "month": int(r["month"]),
        }
        for r in rows
    ]
