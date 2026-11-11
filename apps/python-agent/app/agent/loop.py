import asyncio
import json
from dataclasses import dataclass, field
from typing import Any, Protocol

from mcp import Client
from ollama import AsyncClient

from app.agent.mcp_client import build_server_params
from app.agent.ollama_client import build_ollama_client
from app.agent.tools import mcp_tools_to_ollama
from app.config import load_config

MAX_STEP = 6

SYSTEM_PROMT = (
    "You are a personal expenses assistant. For any numbers or facts, "
    "always call tools and never make up amounts. "
    "If there is no data, say so. Answer briefly and in Russian. "
    "If a list or table is requested, output only a Markdown table. "
    "For questions about spending amounts per month or year, call "
    "get_monthly_totals; get_year_months returns no amounts."
)


@dataclass
class AgentResult:
    answer: str
    used_tools: list[str] = field(default=list)


class McpClient(Protocol):
    async def list_tools(self) -> Any: ...

    async def call_tool(self, name: str, arguments: dict | None = None) -> Any: ...


def tool_result_to_text(result: Any) -> str:
    """CallToolResult -> string used to communicate with the model"""
    structured = getattr(result, "structured_content", None)

    if structured is not None:
        return json.dumps(
            structured,
            ensure_ascii=False,
            default=str,
        )

    parts: list[str] = []

    for block in getattr(result, "content", []) or []:
        text = getattr(block, "text", None)

        if text:
            parts.append(text)

    return "\n".join(parts)


async def run_agent(
        ollama: AsyncClient,
        mcp: McpClient,
        question: str,
        model: str,
        resident_location_id: int,
) -> AgentResult:
    tools_result = await mcp.list_tools()

    tools = mcp_tools_to_ollama(tools_result.tools)

    system_promt = (
        SYSTEM_PROMT + f"\nThe object (resident_location_id): {resident_location_id}"
    )

    messages: list[dict] = [
        { "role": "system", "content": system_promt },
        { "role": "user", "content": question },
    ]

    used: list[str] = []

    for _ in range(MAX_STEP):
        response = await ollama.chat(
            model=model,
            messages=messages,
            tools=tools,
            options={ "temperature": 0.0 },
        )

        msg = response.message

        calls = msg.tool_calls or []

        if not calls:
            return AgentResult(
                answer=msg.content or "",
                used_tools=used,
            )

        messages.append(
            {
                "role": "assistant",
                "content": msg.content or "",
                "tool_calls": calls,
            },
        )

        for call in calls:
            name = call.function.name

            used.append(name)

            try:
                arguments = dict(call.function.arguments or {})

                result = await mcp.call_tool(name, arguments)

                content = tool_result_to_text(result)
            except Exception as exc:
                content = f"Error calling tool {name}: {exc}"

            messages.append(
                {
                    "role": "tool",
                    "tool_name": name,
                    "content": content,
                }
            )

    return AgentResult(
        answer="Not ready to get answer for MAX_STEP",
        used_tools=used,
    )


async def main() -> None:
    cfg = load_config()

    ollama = build_ollama_client(cfg)

    async with Client(build_server_params()) as mcp:
        result = await run_agent(
            ollama=ollama,
            mcp=mcp,
            question="Сколько всего я потратил в июне 2026 года?",
            model=cfg.ollama,
            resident_location_id=1,
        )

    print("Answer:", result.answer)
    print("Used tools:", result.used_tools)


if __name__ == "__main__":
    asyncio.run(main())
