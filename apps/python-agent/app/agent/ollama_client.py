from ollama import AsyncClient

from app.config import Config


def build_ollama_client(cfg: Config) -> AsyncClient:
    """Client ollama; host and model is taken from cfg"""
    return AsyncClient(
        host=cfg.ollama_base_url,
    )

