from psycopg import Connection
from psycopg_pool import ConnectionPool

# Тип пула соединений, возвращаемый create_pool.
Pool = ConnectionPool[Connection]


def create_pool(database_url: str) -> Pool:
    return ConnectionPool(
        conninfo=database_url,
        min_size=1,
        max_size=5,
        timeout=5.0,
        open=True,
    )


