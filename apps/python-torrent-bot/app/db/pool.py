from __future__ import annotations

import logging
import re
from pathlib import Path

from psycopg_pool import AsyncConnectionPool

logger = logging.getLogger(__name__)

_SCHEMA_PATH = Path(__file__).with_name("schema.sql")
# DATABASE_URL содержит пароль — логировать/показывать только
# в отредактированном виде (design §Error Handling)
_PASSWORD_RE = re.compile(r"://([^:/@]+):[^@/]+@")


def redact_database_url(url: str) -> str:
    return _PASSWORD_RE.sub(r"://\1:***@", url)


async def create_pool(database_url: str) -> AsyncConnectionPool:
    pool = AsyncConnectionPool(
        database_url,
        min_size=1,
        max_size=5,
        open=False,
        kwargs={"connect_timeout": 5},
    )
    try:
        await pool.open(wait=True, timeout=10)
    except Exception as exc:
        await pool.close()
        logger.error("cannot reach PostgreSQL at %s: %s", redact_database_url(database_url), exc)
        raise RuntimeError(
            f"Cannot reach PostgreSQL at {redact_database_url(database_url)}: {exc}"
        ) from exc
    logger.info("database pool opened")
    return pool


async def apply_schema(pool: AsyncConnectionPool) -> None:
    script = _SCHEMA_PATH.read_text(encoding="utf-8")
    # строки-комментарии вырезаем до split по ';': комментарии schema.sql содержат ';'
    code = "\n".join(ln for ln in script.splitlines() if not ln.lstrip().startswith("--"))
    statements = [s.strip() for s in code.split(";") if s.strip()]
    try:
        async with pool.connection() as conn:
            for statement in statements:
                await conn.execute(statement)
    except Exception as exc:
        # URL уже прошёл проверку в create_pool; в сообщении его нет — пароль не утекает
        logger.error("cannot apply database schema: %s", exc)
        raise RuntimeError(f"Cannot apply database schema: {exc}") from exc
    logger.info("database schema applied")
