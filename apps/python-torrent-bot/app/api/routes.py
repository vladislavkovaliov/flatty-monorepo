from __future__ import annotations

import logging
from collections.abc import Awaitable

from fastapi import APIRouter, HTTPException, status

from app.api.schemas import ErrorOut, TaskCreateIn, TorrentTaskOut
from app.models.torrent import TorrentStatus, TorrentTask
from app.repositories import TorrentRepository
from app.services.magnet_parser import parse_magnet
from app.services.torrent_service import TorrentService

logger = logging.getLogger(__name__)

_PAUSEABLE = (TorrentStatus.QUEUED, TorrentStatus.DOWNLOADING)


async def _guard[T](op: str, coro: Awaitable[T]) -> T:
    """Неожиданная ошибка сервиса -> log + 500 (design §Error Handling)."""
    try:
        return await coro
    except Exception:
        logger.error("api %s failed", op, exc_info=True)
        raise HTTPException(status_code=500, detail="internal error") from None


def create_router(
    service: TorrentService, repository: TorrentRepository, api_user_id: int
) -> APIRouter:
    router = APIRouter(prefix="/api/tasks", tags=["tasks"])

    def _out(task: TorrentTask) -> TorrentTaskOut:
        return TorrentTaskOut.model_validate(task)

    async def _owned(task_id: int) -> TorrentTask:
        task = await _guard("get", service.get(api_user_id, task_id))
        if task is None:
            raise HTTPException(status_code=404, detail="task not found")
        return task

    async def _mutated(op: str, task: TorrentTask | None) -> TorrentTaskOut:
        if task is None:
            raise HTTPException(status_code=404, detail="task not found")
        return _out(task)

    @router.get("", response_model=list[TorrentTaskOut])
    async def list_tasks() -> list[TorrentTaskOut]:
        tasks = await _guard("list", service.list(api_user_id))
        return [_out(task) for task in tasks]

    @router.get(
        "/{task_id}",
        response_model=TorrentTaskOut,
        responses={404: {"model": ErrorOut}},
    )
    async def get_task(task_id: int) -> TorrentTaskOut:
        return _out(await _owned(task_id))

    @router.post(
        "",
        status_code=status.HTTP_201_CREATED,
        response_model=TorrentTaskOut,
        responses={
            400: {"model": ErrorOut},
            409: {"description": "Duplicate; detail содержит существующую задачу"},
        },
    )
    async def create_task(payload: TaskCreateIn) -> TorrentTaskOut:
        try:
            magnet = parse_magnet(payload.magnet)
        except ValueError as exc:
            raise HTTPException(status_code=400, detail=str(exc)) from None
        if not await _guard("user_exists", repository.user_exists(api_user_id)):
            raise HTTPException(
                status_code=400, detail="пользователь не найден — нажми /start в боте"
            )
        task, created = await _guard(
            "create",
            service.create(
                user_id=api_user_id,
                magnet=magnet,
                media_type=payload.media_type,
                quality=payload.quality,
            ),
        )
        if not created:
            raise HTTPException(status_code=409, detail=_out(task).model_dump(mode="json"))
        return _out(task)

    @router.post(
        "/{task_id}/pause",
        response_model=TorrentTaskOut,
        responses={404: {"model": ErrorOut}, 409: {"model": ErrorOut}},
    )
    async def pause_task(task_id: int) -> TorrentTaskOut:
        task = await _owned(task_id)
        # сервис при неверном статусе вернул бы задачу без изменений — ловим предпроверкой
        if task.status not in _PAUSEABLE:
            raise HTTPException(
                status_code=409, detail=f"cannot pause task in status {task.status}"
            )
        return await _mutated("pause", await _guard("pause", service.pause(api_user_id, task_id)))

    @router.post(
        "/{task_id}/resume",
        response_model=TorrentTaskOut,
        responses={404: {"model": ErrorOut}, 409: {"model": ErrorOut}},
    )
    async def resume_task(task_id: int) -> TorrentTaskOut:
        task = await _owned(task_id)
        if task.status is not TorrentStatus.PAUSED:
            raise HTTPException(
                status_code=409, detail=f"cannot resume task in status {task.status}"
            )
        return await _mutated(
            "resume", await _guard("resume", service.resume(api_user_id, task_id))
        )

    @router.post(
        "/{task_id}/cancel",
        response_model=TorrentTaskOut,
        responses={404: {"model": ErrorOut}, 409: {"model": ErrorOut}},
    )
    async def cancel_task(task_id: int) -> TorrentTaskOut:
        task = await _owned(task_id)
        # сервис cancel не отклоняет никогда — 409 только на уже отменённую (design §Components)
        if task.status is TorrentStatus.CANCELLED:
            raise HTTPException(status_code=409, detail="task already cancelled")
        return await _mutated(
            "cancel", await _guard("cancel", service.cancel(api_user_id, task_id))
        )

    @router.delete(
        "/{task_id}",
        status_code=status.HTTP_204_NO_CONTENT,
        responses={404: {"model": ErrorOut}},
    )
    async def delete_task(task_id: int) -> None:
        removed = await _guard("delete", service.delete(api_user_id, task_id))
        if not removed:
            # False = «не найдено»; сбой provider тоже сюда (ponytail: сервис различить не даёт)
            raise HTTPException(status_code=404, detail="task not found")

    return router
