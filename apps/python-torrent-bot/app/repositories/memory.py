from __future__ import annotations

import itertools

from app.models.torrent import TorrentTask


class InMemoryRepository:
    """Тестовый double: тот же контракт, что и PostgresRepository, без инфраструктуры."""

    def __init__(self) -> None:
        self._tasks: dict[int, TorrentTask] = {}
        self._ids = itertools.count(1)
        self.events: list[tuple[int, str, str | None]] = []
        self._settings: dict[tuple[int, str], str] = {}
        self._users: set[int] = set()

    async def find_by_hash(self, user_id: int, info_hash: str) -> TorrentTask | None:
        for task in self._tasks.values():
            if task.user_id == user_id and task.info_hash == info_hash:
                return task
        return None

    async def insert_task(self, task: TorrentTask) -> TorrentTask | None:
        if await self.find_by_hash(task.user_id, task.info_hash) is not None:
            return None  # как ON CONFLICT (user_id, info_hash) DO NOTHING
        task.id = next(self._ids)
        self._tasks[task.id] = task
        self._users.add(task.user_id)
        return task

    async def get_task(self, user_id: int, task_id: int) -> TorrentTask | None:
        task = self._tasks.get(task_id)
        if task is None or task.user_id != user_id:
            return None
        return task

    async def list_tasks(self, user_id: int) -> list[TorrentTask]:
        return [t for t in self._tasks.values() if t.user_id == user_id]

    async def update_task(self, task: TorrentTask) -> None:
        # объекты хранятся по ссылке — мутации уже видны; совпадает с контрактом «персист»
        return None

    async def delete_task(self, task_id: int) -> None:
        self._tasks.pop(task_id, None)

    async def add_event(self, torrent_id: int, event: str, payload: str | None = None) -> None:
        self.events.append((torrent_id, event, payload))

    async def get_setting(self, user_id: int, key: str) -> str | None:
        return self._settings.get((user_id, key))

    async def set_setting(self, user_id: int, key: str, value: str) -> None:
        self._settings[(user_id, key)] = value
        self._users.add(user_id)

    async def ensure_user(self, telegram_user_id: int) -> None:
        self._users.add(telegram_user_id)

    async def user_exists(self, telegram_user_id: int) -> bool:
        return telegram_user_id in self._users
