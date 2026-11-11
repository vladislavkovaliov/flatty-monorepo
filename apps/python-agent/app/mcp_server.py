from mcp.server import MCPServer

from app.config import load_config
from app.db import expenses as repo
from app.db.connection import Pool, create_pool
from app.db.types import MonthlyTotalRow, CategoryRow


def build_mcp_server(pool: Pool) -> MCPServer:
    mcp = MCPServer("expense-agent")

    @mcp.tool()
    def get_monthly_totals(
        resident_location_id: int,
        year: int | None = None,
        month: int | None = None
    ) -> list[MonthlyTotalRow]:
        """Вернуть суммы трат по месяцам для объекта."""

        return repo.get_monthly_totals(
            pool=pool,
            resident_location_id=resident_location_id,
            year=year,
            month=month,
        )

    @mcp.resource(uri="resource://categories")
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
