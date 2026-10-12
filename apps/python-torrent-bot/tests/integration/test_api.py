from __future__ import annotations

from collections.abc import AsyncGenerator

import pytest
from fastapi.testclient import TestClient

from app.api.app import create_api_app
from app.providers.fake import FakeTorrentProvider
from app.repositories.memory import InMemoryRepository
from app.services.torrent_service import TorrentService

_API_USER = 1
_MAGNET = "magnet:?xt=urn:btih:" + "ab" * 20 + "&dn=Test.Torrent"
_MAGNET_OTHER = "magnet:?xt=urn:btih:" + "cd" * 20 + "&dn=Other"


@pytest.fixture
async def repo() -> AsyncGenerator[InMemoryRepository, None]:
    r = InMemoryRepository()
    await r.ensure_user(_API_USER)  # «пользователь нажал /start»
    yield r


@pytest.fixture
async def client(repo: InMemoryRepository) -> AsyncGenerator[TestClient, None]:
    service = TorrentService(FakeTorrentProvider(), repo)
    with TestClient(create_api_app(service, repo, _API_USER)) as c:
        yield c


async def test_create_list_flow(client: TestClient) -> None:
    created = client.post(
        "/api/tasks",
        json={"magnet": _MAGNET, "media_type": "movie", "quality": "1080p"},
    )
    assert created.status_code == 201
    body = created.json()
    assert body["id"] == 1
    assert body["user_id"] == _API_USER
    assert body["status"] == "queued"  # StrEnum -> plain string в JSON
    assert body["media_type"] == "movie"
    assert body["quality"] == "1080p"

    listed = client.get("/api/tasks")
    assert listed.status_code == 200
    assert [t["id"] for t in listed.json()] == [1]

    single = client.get("/api/tasks/1")
    assert single.status_code == 200
    assert single.json()["name"] == "Test.Torrent"


async def test_pause_resume_cancel_delete_flow(client: TestClient) -> None:
    task_id = client.post(
        "/api/tasks",
        json={"magnet": _MAGNET, "media_type": "movie", "quality": "1080p"},
    ).json()["id"]

    paused = client.post(f"/api/tasks/{task_id}/pause")
    assert paused.status_code == 200
    assert paused.json()["status"] == "paused"

    resumed = client.post(f"/api/tasks/{task_id}/resume")
    assert resumed.status_code == 200
    assert resumed.json()["status"] == "downloading"

    cancelled = client.post(f"/api/tasks/{task_id}/cancel")
    assert cancelled.status_code == 200
    assert cancelled.json()["status"] == "cancelled"

    deleted = client.delete(f"/api/tasks/{task_id}")
    assert deleted.status_code == 204
    assert deleted.content == b""

    assert client.get(f"/api/tasks/{task_id}").status_code == 404


async def test_duplicate_returns_409_with_existing_task(client: TestClient) -> None:
    first = client.post(
        "/api/tasks",
        json={"magnet": _MAGNET, "media_type": "movie", "quality": "1080p"},
    )
    assert first.status_code == 201

    duplicate = client.post(
        "/api/tasks",
        json={"magnet": _MAGNET, "media_type": "series", "quality": "720p"},
    )
    assert duplicate.status_code == 409
    detail = duplicate.json()["detail"]
    assert detail["id"] == first.json()["id"]
    assert detail["media_type"] == "movie"  # существующая задача, а не новая


async def test_wrong_state_mutations_409(client: TestClient) -> None:
    task_id = client.post(
        "/api/tasks",
        json={"magnet": _MAGNET, "media_type": "movie", "quality": "1080p"},
    ).json()["id"]

    assert client.post(f"/api/tasks/{task_id}/pause").status_code == 200
    assert client.post(f"/api/tasks/{task_id}/pause").status_code == 409  # уже на паузе

    assert client.post(f"/api/tasks/{task_id}/resume").status_code == 200
    assert client.post(f"/api/tasks/{task_id}/resume").status_code == 409  # не на паузе

    assert client.post(f"/api/tasks/{task_id}/cancel").status_code == 200
    assert client.post(f"/api/tasks/{task_id}/cancel").status_code == 409  # уже отменена


async def test_get_and_delete_404(client: TestClient) -> None:
    assert client.get("/api/tasks/999").status_code == 404
    assert client.delete("/api/tasks/999").status_code == 404


async def test_create_422_validation(client: TestClient) -> None:
    invalid_enum = client.post(
        "/api/tasks",
        json={"magnet": _MAGNET, "media_type": "book", "quality": "1080p"},
    )
    assert invalid_enum.status_code == 422

    missing_field = client.post("/api/tasks", json={"magnet": _MAGNET})
    assert missing_field.status_code == 422


async def test_create_400_bad_magnet(client: TestClient) -> None:
    response = client.post(
        "/api/tasks",
        json={
            "magnet": "http://example.com/file.torrent",
            "media_type": "movie",
            "quality": "1080p",
        },
    )
    assert response.status_code == 400
    assert "scheme" in response.json()["detail"]


async def test_create_400_when_api_user_unknown() -> None:
    repo = InMemoryRepository()  # никто не нажал /start
    service = TorrentService(FakeTorrentProvider(), repo)
    client = TestClient(create_api_app(service, repo, _API_USER))
    response = client.post(
        "/api/tasks",
        json={"magnet": _MAGNET, "media_type": "movie", "quality": "1080p"},
    )
    assert response.status_code == 400
    assert "/start" in response.json()["detail"]


async def test_openapi_and_docs(client: TestClient) -> None:
    schema = client.get("/openapi.json")
    assert schema.status_code == 200
    assert "/api/tasks" in schema.json()["paths"]
    assert client.get("/docs").status_code == 200
