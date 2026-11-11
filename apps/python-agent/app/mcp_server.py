from mcp.server import MCPServer

from app.config import load_config
from app.db import expenses as repo
from app.db.connection import Pool, create_pool
from app.db.types import CategoryRow, ExpenseRow, MonthlyAverageRow, MonthlyTotalRow, YearMonthRow


def build_mcp_server(pool: Pool) -> MCPServer:
    mcp = MCPServer("expense-agent")

    @mcp.tool()
    def get_year_months(
        resident_location_id: int
    ) -> list[YearMonthRow]:
        """Return the (year, month) pairs that have expenses."""
        return repo.get_year_months(
            pool=pool,
            resident_location_id=resident_location_id
        )

    @mcp.tool()
    def get_month_averages(
        resident_location_id: int
    ) -> list[MonthlyAverageRow]:
        """Return monthly average spending."""
        return repo.get_month_averages(
            pool=pool,
            resident_location_id=resident_location_id
        )

    @mcp.tool()
    def get_monthly_totals(
        resident_location_id: int,
        year: int | None = None,
        month: int | None = None
    ) -> list[MonthlyTotalRow]:
        """Return monthly spending totals for a location."""
        return repo.get_monthly_totals(
            pool=pool,
            resident_location_id=resident_location_id,
            year=year,
            month=month,
        )

    @mcp.tool()
    def list_expenses(
        resident_location_id: int,
        category_id: int | None = None,
        month: int | None = None,
        year: int | None = None,
        limit: int = 5,
        offset: int = 0
    ) -> list[ExpenseRow]:
        """Return the list of expenses."""
        return repo.list_expenses(
            pool=pool,
            resident_location_id=resident_location_id,
            limit=limit,
            month=month,
            year=year,
            offset=offset,
            category_id=category_id
        )

    @mcp.tool()
    def list_categories() -> list[CategoryRow]:
        """Return the list of expense categories (id, name, description)."""
        return repo.list_categories(pool)

    @mcp.resource(
        uri="resource://categories",
        name="categories",
        description="Expense categories reference (id, name, description).",
    )
    def categories() -> list[CategoryRow]:
        return repo.list_categories(pool)

    return mcp



if __name__ == "__main__":
    cfg = load_config()
    pool = create_pool(cfg.database_url)
    mcp = build_mcp_server(pool)

    try:
        mcp.run()
    finally:
        pool.close()
