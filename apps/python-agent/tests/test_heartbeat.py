import asyncio

from app.api.app import sse_with_heartbeat


class FakeRequest:
    async def is_disconnected(self) -> bool:
        return False


class DisconnectedRequest:
    async def is_disconnected(self) -> bool:
        return True


async def slow_events():
    await asyncio.sleep(0.05)
    yield {"event": "done", "data": {"answer": "ok"}}


def test_heartbeat_pings_while_waiting(monkeypatch):
    monkeypatch.setattr("app.api.app.HEARTBEAT_SECONDS", 0.01)

    async def collect():
        return [c async for c in sse_with_heartbeat(slow_events(), FakeRequest())]

    chunks = asyncio.run(collect())

    assert chunks[-1].startswith("event: done")
    assert chunks[:-1] and all(c == ": ping\n\n" for c in chunks[:-1])


def test_disconnected_client_stops_stream(monkeypatch):
    monkeypatch.setattr("app.api.app.HEARTBEAT_SECONDS", 0.01)

    async def collect():
        return [c async for c in sse_with_heartbeat(slow_events(), DisconnectedRequest())]

    assert asyncio.run(collect()) == []
