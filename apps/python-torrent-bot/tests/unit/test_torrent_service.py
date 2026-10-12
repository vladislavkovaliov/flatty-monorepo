from __future__ import annotations

import pytest

from app.models.media import MediaType, Quality
from app.models.torrent import TorrentStatus, TorrentTask
from app.providers.fake import FakeTorrentProvider
from app.repositories.memory import InMemoryRepository
from app.services.magnet_parser import MagnetInfo
from app.services.torrent_service import TorrentService

HASH_A = "a" * 40
HASH_B = "b1" + "0" * 38


def _magnet(info_hash: str = HASH_A, name: str | None = "The Matrix") -> MagnetInfo:
    raw = f"magnet:?xt=urn:btih:{info_hash}" + (f"&dn={name}" if name else "")
    return MagnetInfo(
        raw=raw,
        info_hash=info_hash.lower(),
        display_name=name,
        trackers=[],
        size_bytes=None,
    )


@pytest.fixture
def repo() -> InMemoryRepository:
    return InMemoryRepository()


@pytest.fixture
def service(repo: InMemoryRepository) -> TorrentService:
    return TorrentService(FakeTorrentProvider(), repo)


async def test_create_makes_queued_task(service: TorrentService) -> None:
    task, created = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    assert created is True
    assert task.status is TorrentStatus.QUEUED
    assert task.user_id == 1
    assert task.name == "The Matrix"
    assert task.provider_id.startswith("fake-")


async def test_duplicate_by_info_hash_same_user(service: TorrentService) -> None:
    first, created_first = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    second, created_second = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.SERIES, quality=Quality.HD_720
    )
    assert created_first is True
    assert created_second is False
    assert second.id == first.id


async def test_different_info_hash_creates_separate_tasks(service: TorrentService) -> None:
    first, _ = await service.create(
        user_id=1, magnet=_magnet(HASH_A), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    second, created = await service.create(
        user_id=1, magnet=_magnet(HASH_B), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    assert created is True
    assert second.id != first.id
    assert len(await service.list(1)) == 2


async def test_other_user_can_create_same_hash(service: TorrentService) -> None:
    # FEAT-009 + интерпретация №3: дубль scope = (user_id, info_hash)
    mine, _ = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    other, created = await service.create(
        user_id=2, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    assert created is True
    assert other.id != mine.id


async def test_user_isolation_reads(service: TorrentService) -> None:
    task, _ = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    assert await service.list(2) == []
    assert await service.get(2, task.id) is None
    assert await service.get(1, task.id) is not None


async def test_ownership_enforced_on_operations(service: TorrentService) -> None:
    task, _ = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    assert await service.pause(2, task.id) is None
    assert await service.resume(2, task.id) is None
    assert await service.cancel(2, task.id) is None
    assert await service.delete(2, task.id) is False
    untouched = await service.get(1, task.id)
    assert untouched is not None
    assert untouched.status is TorrentStatus.QUEUED


async def test_status_transitions(service: TorrentService) -> None:
    task, _ = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    assert task.status is TorrentStatus.QUEUED

    for _ in range(3):
        task = (await service.list(1))[0]
        if task.status is not TorrentStatus.QUEUED:
            break
    assert task.status is TorrentStatus.DOWNLOADING

    paused = await service.pause(1, task.id)
    assert paused is not None and paused.status is TorrentStatus.PAUSED
    assert (await service.list(1))[0].status is TorrentStatus.PAUSED

    resumed = await service.resume(1, task.id)
    assert resumed is not None and resumed.status is TorrentStatus.DOWNLOADING

    cancelled = await service.cancel(1, task.id)
    assert cancelled is not None and cancelled.status is TorrentStatus.CANCELLED
    assert (await service.list(1))[0].status is TorrentStatus.CANCELLED

    assert await service.delete(1, task.id) is True
    assert await service.list(1) == []


async def test_progress_reaches_completed(service: TorrentService) -> None:
    task, _ = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    for _ in range(20):
        task = (await service.list(1))[0]
        if task.status is TorrentStatus.COMPLETED:
            break
    assert task.status is TorrentStatus.COMPLETED
    assert task.progress == 100.0
    assert task.completed_at is not None


async def test_provider_reregistered_after_restart(repo: InMemoryRepository) -> None:
    # «рестарт»: задача уже в хранилище, новый сервис + новый (пустой) provider
    seeded = TorrentTask(
        id=0,
        user_id=1,
        magnet=f"magnet:?xt=urn:btih:{HASH_A}",
        info_hash=HASH_A,
        name="Restored",
        media_type=MediaType.MOVIE,
        quality=Quality.FHD_1080,
    )
    assert await repo.insert_task(seeded) is not None
    service = TorrentService(FakeTorrentProvider(), repo)
    tasks = await service.list(1)
    assert [t.name for t in tasks] == ["Restored"]
    assert tasks[0].provider_id.startswith("fake-")


async def test_status_changes_record_lifecycle_events(
    service: TorrentService, repo: InMemoryRepository
) -> None:
    task, _ = await service.create(
        user_id=1, magnet=_magnet(), media_type=MediaType.MOVIE, quality=Quality.FHD_1080
    )
    await service.pause(1, task.id)
    await service.cancel(1, task.id)
    await service.delete(1, task.id)
    kinds = [event for _tid, event, _payload in repo.events]
    assert kinds == ["created", "status_changed", "status_changed", "deleted"]
