import uvicorn

from app.config import load_config

if __name__ == "__main__":
    cfg = load_config()
    print(42)

    uvicorn.run(
        "app.api.app:app",
        host=cfg.api_host,
        port=cfg.api_port,    
    )