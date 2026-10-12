from __future__ import annotations

from fastapi import FastAPI

from app.api.routes import create_router
from app.repositories import TorrentRepository
from app.services.torrent_service import TorrentService


def create_api_app(
    service: TorrentService, repository: TorrentRepository, api_user_id: int
) -> FastAPI:
    app = FastAPI(
        title="Torrent Download Bot API",
        version="0.1.0",
        docs_url="/docs",
        openapi_url="/openapi.json",
    )
    app.include_router(create_router(service, repository, api_user_id))
    return app
