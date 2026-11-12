from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, Field
from typing import Literal

from app.agent.loop import run_agent
from app.agent.mcp_client import build_server_params
from app.agent.ollama_client import build_ollama_client
from app.config import load_config

from mcp import Client


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

def trim_history(history: list[HistoryMessage]) -> list[dict]:
    tail = history[-MAX_HISTORY_MESSAGES:]

    budget = MAX_HISTORY_CHARS
    ketp: list[dict] = []

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


@asynccontextmanager
async def lifespan(app: FastAPI):
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
        )

    return ChatResponse(
        answer=result.answer,
        used_tools=result.used_tools,
    )