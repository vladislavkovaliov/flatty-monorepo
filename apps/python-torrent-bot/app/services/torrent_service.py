from __future__ import annotations

import json
import logging
from datetime import UTC, datetime

from app.models.media import MediaType, Quality
from app.models.torrent import TorrentStatus, TorrentTask
from app.providers.base import TorrentProvider
from app.repositories import TorrentRepository
from app.services.magnet_parser import MagnetInfo

logger = logging.getLogger(__name__)

_DEFAULT_SAVE_PATHS: dict[MediaType, str] = {
    MediaType.MOVIE: "Movies/",
    MediaType.SERIES: "TV/",
}
# PAUSED/COMPLETED/FAILED/CANCELLED не опрашиваются у provider (нет ложных переходов)
_TERMINAL = frozenset(
    {
        TorrentStatus.PAUSED,
        TorrentStatus.COMPLETED,
        TorrentStatus.FAILED,
        TorrentStatus.CANCELLED,
    }
)


class TorrentService:
    """Бизнес-логика очереди: валидация, duplicate по info_hash, ownership, вызовы provider.

    Хранение — за TorrentRepository (design §Architecture); handlers/Telegram не знают о БД.
    Без Telegram-кода (docs/API.md §3).
    """

    def __init__(
        self,
        provider: TorrentProvider,
        repository: TorrentRepository,
        save_paths: dict[MediaType, str] | None = None,
    ) -> None:
        self._provider = provider
        self._repo = repository
        self._save_paths = save_paths or dict(_DEFAULT_SAVE_PATHS)

    def save_path_for(self, media_type: MediaType) -> str:
        return self._save_paths[media_type]

    async def register_user(self, user_id: int) -> None:
        """Регистрация при /start: строка bot_users нужна API до первого create (FEAT-020)."""
        await self._repo.ensure_user(user_id)

    async def create(
        self,
        *,
        user_id: int,
        magnet: MagnetInfo,
        media_type: MediaType,
        quality: Quality,
    ) -> tuple[TorrentTask, bool]:
        existing = await self._repo.find_by_hash(user_id, magnet.info_hash)
        if existing is not None:
            logger.info(
                "duplicate info_hash=%s user_id=%s task_id=%s",
                magnet.info_hash,
                user_id,
                existing.id,
            )
            await self._sync(existing)
            return existing, False

        task = TorrentTask(
            id=0,
            user_id=user_id,
            magnet=magnet.raw,
            info_hash=magnet.info_hash,
            name=magnet.display_name or f"magnet-{magnet.info_hash[:8]}",
            media_type=media_type,
            quality=quality,
            size_bytes=magnet.size_bytes,
        )
        inserted = await self._repo.insert_task(task)
        if inserted is None:
            # гонка на UNIQUE(user_id, info_hash): кто-то вставил между find и insert
            existing = await self._repo.find_by_hash(user_id, magnet.info_hash)
            if existing is None:
                raise RuntimeError(f"torrent conflict resolved but row missing: {magnet.info_hash}")
            logger.info(
                "duplicate info_hash=%s user_id=%s task_id=%s",
                magnet.info_hash,
                user_id,
                existing.id,
            )
            return existing, False
        task = inserted

        try:
            await self._ensure_provider(task)
        except Exception:
            logger.error("provider.add_magnet failed task_id=%s", task.id, exc_info=True)
            await self._change(task, TorrentStatus.FAILED)
            return task, True

        await self._repo.add_event(task.id, "created")
        logger.info(
            "created task_id=%s user_id=%s info_hash=%s", task.id, user_id, magnet.info_hash
        )
        return task, True

    async def list(self, user_id: int) -> list[TorrentTask]:
        tasks = await self._repo.list_tasks(user_id)
        for task in tasks:
            await self._sync(task)
        return tasks

    async def list_completed_tasks(self, user_id: int) -> int:
        count = await self._repo.list_completed_count_tasks(user_id)
        return count

    async def list_downloading_tasks(self, user_id: int) -> int:
        count = await self._repo.list_downloading_count_tasks(user_id)
        return count

    async def list_queue_count_tasks(self, user_id: int) -> int:
        count = await self._repo.list_queue_count_tasks(user_id)
        return count


    async def get(self, user_id: int, task_id: int) -> TorrentTask | None:
        return await self._owned(user_id, task_id)

    async def pause(self, user_id: int, task_id: int) -> TorrentTask | None:
        task = await self._owned(user_id, task_id)
        if task is None or task.status not in (
            TorrentStatus.QUEUED,
            TorrentStatus.DOWNLOADING,
        ):
            return task
        try:
            await self._provider.pause(task.provider_id)
        except Exception:
            logger.error("provider.pause failed task_id=%s", task.id, exc_info=True)
            await self._change(task, TorrentStatus.FAILED)
            return task
        await self._change(task, TorrentStatus.PAUSED)
        logger.info("paused task_id=%s user_id=%s", task_id, user_id)
        return task

    async def resume(self, user_id: int, task_id: int) -> TorrentTask | None:
        task = await self._owned(user_id, task_id)
        if task is None or task.status is not TorrentStatus.PAUSED:
            return task
        try:
            await self._provider.resume(task.provider_id)
        except Exception:
            logger.error("provider.resume failed task_id=%s", task.id, exc_info=True)
            await self._change(task, TorrentStatus.FAILED)
            return task
        await self._change(task, TorrentStatus.DOWNLOADING)
        logger.info("resumed task_id=%s user_id=%s", task_id, user_id)
        return task

    async def cancel(self, user_id: int, task_id: int) -> TorrentTask | None:
        task = await self._owned(user_id, task_id)
        if task is None:
            return None
        if task.provider_id:
            try:
                await self._provider.delete(task.provider_id, delete_files=False)
            except Exception:
                logger.error("provider.delete failed task_id=%s", task.id, exc_info=True)
        await self._change(task, TorrentStatus.CANCELLED)
        logger.info("cancelled task_id=%s user_id=%s", task_id, user_id)
        return task

    async def delete(self, user_id: int, task_id: int) -> bool:
        task = await self._owned(user_id, task_id)
        if task is None:
            return False
        try:
            await self._provider.delete(task.provider_id)
        except Exception:
            logger.error("provider.delete failed task_id=%s", task.id, exc_info=True)
            return False
        # событие ДО удаления: torrent_events.torrent_id — FK с ON DELETE CASCADE
        await self._repo.add_event(task.id, "deleted")
        await self._repo.delete_task(task_id)
        logger.info("deleted task_id=%s user_id=%s", task_id, user_id)
        return True

    async def _owned(self, user_id: int, task_id: int) -> TorrentTask | None:
        task = await self._repo.get_task(user_id, task_id)
        if task is None:
            # FEAT-009: чужая/несуществующая — трактовать как «не найдено»
            return None
        await self._sync(task)
        return task

    async def _ensure_provider(self, task: TorrentTask) -> None:
        """Регистрирует magnet у provider и персистит provider_id (FEAT-011).

        После рестарта задача с пустым provider_id получает magnet заново
        (дедуп по info_hash на стороне провайдера).
        """
        if task.provider_id:
            return
        task.provider_id = await self._provider.add_magnet(
            task.magnet, self._save_paths[task.media_type]
        )
        await self._repo.update_task(task)

    async def _change(self, task: TorrentTask, new_status: TorrentStatus) -> None:
        if new_status is not task.status:
            await self._repo.add_event(
                task.id,
                "status_changed",
                json.dumps({"from": task.status, "to": new_status}),
            )
        task.status = new_status
        await self._repo.update_task(task)

    async def _sync(self, task: TorrentTask) -> None:
        if task.status in _TERMINAL:
            return
        try:
            await self._ensure_provider(task)
        except Exception:
            logger.error("provider.add_magnet failed task_id=%s", task.id, exc_info=True)
            await self._change(task, TorrentStatus.FAILED)
            return
        try:
            info = await self._provider.get_info(task.provider_id)
        except Exception:
            logger.error("provider.get_info failed task_id=%s", task.id, exc_info=True)
            await self._change(task, TorrentStatus.FAILED)
            return
        if info.status is not task.status:
            # lifecycle-событие только на смене статуса, НЕ на тике прогресса (design §Components)
            await self._repo.add_event(
                task.id,
                "status_changed",
                json.dumps({"from": task.status, "to": info.status}),
            )
            task.status = info.status
        task.progress = info.progress
        task.download_speed = info.download_speed
        task.upload_speed = info.upload_speed
        task.eta_seconds = info.eta_seconds
        if info.status is TorrentStatus.COMPLETED and task.completed_at is None:
            task.completed_at = datetime.now(UTC)
        await self._repo.update_task(task)
