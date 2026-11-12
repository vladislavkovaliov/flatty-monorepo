import asyncio
import json
from collections.abc import AsyncIterator
from contextlib import asynccontextmanager
from typing import Any, Literal

from fastapi import FastAPI, HTTPException, Request
from fastapi.responses import StreamingResponse
from mcp import Client
from pydantic import BaseModel, Field

from app.agent.loop import run_agent, run_agent_stream
from app.agent.mcp_client import build_server_params
from app.agent.ollama_client import build_ollama_client
from app.config import load_config


class HistoryMessage(BaseModel):
    role: Literal["user", "assistant"]
    content: str


class ChatRequest(BaseModel):
    message: str
    resident_location_id: int = 1
    history: list[HistoryMessage] = Field(default_factory=list)

class ChatResponse(BaseModel):
    answer: str
    used_tools: list[str]


MAX_HISTORY_MESSAGES = 20
MAX_HISTORY_CHARS = 8000

# INFO: keepalive interval, well under typical 15-20s proxy/browser idle timeouts
HEARTBEAT_SECONDS = 10.0

def trim_history(history: list[HistoryMessage]) -> list[dict[str, Any]]:
    tail = history[-MAX_HISTORY_MESSAGES:]

    budget = MAX_HISTORY_CHARS
    ketp: list[dict[str, Any]] = []

    for msg in reversed(tail):
        if len(msg.content) > budget:
            break

        budget = budget - len(msg.content)
        ketp.append({
            "role": msg.role,
            "content": msg.content,
        })

    ketp.reverse()

    # INFO: don't start history answer w'o question
    if ketp and ketp[0]["role"] == "assistant":
        ketp = ketp[1:]

    return ketp


def sse(event: str, data: dict[str, Any]) -> str:
    return f"event: {event}\ndata: {json.dumps(data, ensure_ascii=False)}\n\n"


async def sse_with_heartbeat(
    events: AsyncIterator[dict[str, Any]],
    request: Request,
) -> AsyncIterator[str]:
    while True:
        pending = asyncio.ensure_future(anext(events))

        while True:
            done, _ = await asyncio.wait({pending}, timeout=HEARTBEAT_SECONDS)

            if done:
                break

            if await request.is_disconnected():
                pending.cancel()
                return

            # INFO: comment frame — ignored by SSE parsers, keeps connection alive
            yield ": ping\n\n"

        try:
            ev = pending.result()
        except StopAsyncIteration:
            return

        if await request.is_disconnected():
            return

        yield sse(ev["event"], ev["data"])



@asynccontextmanager
async def lifespan(app: FastAPI) -> AsyncIterator[None]:
    cfg = load_config()

    app.state.model = cfg.ollama
    app.state.ollama = build_ollama_client(cfg=cfg)

    async with Client(build_server_params()) as mcp:
        app.state.mcp = mcp

        yield


app = FastAPI(
    title="agent",
    lifespan=lifespan,
)


@app.get('/health')
def health() -> dict[str, str]:
    return {
        "status": "ok",
    }


@app.post('/chat')
async def chat(req: ChatRequest) -> ChatResponse:
    if not req.message.strip():
        raise HTTPException(
            status_code=400,
            detail="message is empty",
        )

    try:
        result = await run_agent(
            app.state.ollama,
            app.state.mcp,
            req.message,
            app.state.model,
            req.resident_location_id,
            history=trim_history(req.history),
        )
    except Exception as exc:
        raise HTTPException(
            status_code=503,
            detail=f"agent error: {exc}",
        ) from exc

    return ChatResponse(
        answer=result.answer,
        used_tools=result.used_tools,
    )


@app.post('/chat/stream')
async def chat_stream(
    req: ChatRequest,
    request: Request,
) -> StreamingResponse:
    if not req.message.strip():
        raise HTTPException(
            status_code=400,
            detail="message is empty",
        )

    async def event_source() -> AsyncIterator[str]:
        try:
            async for chunk in sse_with_heartbeat(
                run_agent_stream(
                    app.state.ollama,
                    app.state.mcp,
                    req.message,
                    app.state.model,
                    resident_location_id=req.resident_location_id,
                    history=trim_history(req.history),
                ),
                request,
            ):
                yield chunk
        except Exception as exc:
            yield sse("error", { "detail": str(exc) })

    return StreamingResponse(
        event_source(),
        media_type="text/event-stream",
        headers={
            "Cache-Control": "no-cache",
            "X-Accel-Buffering": "no",
            "Connection": "keep-alive",
        }
    )