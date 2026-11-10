from psycopg import Connection
from psycopg_pool import ConnectionPool


def create_pool(database_url: str) -> "ConnectionPool[Connection]":
    return ConnectionPool(
        conninfo=database_url,
        min_size=1,
        max_size=5,
        timeout=5.0,
        open=True,
    )


