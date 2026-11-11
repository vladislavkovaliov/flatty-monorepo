from contextlib import asynccontextmanager

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel

from app.agent.loop import run_agent
from app.agent.mcp_client import build_server_params
from app.agent.ollama_client import build_ollama_client
from app.config import load_config

from mcp import Client


class ChatRequest(BaseModel):
    message: str
    resident_location_id: int = 1


class ChatResponse(BaseModel):
    answer: str
    used_tools: list[str]


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