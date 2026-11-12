import asyncio
import json
from collections.abc import AsyncGenerator
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
    "get_monthly_totals; get_year_months returns no amounts. "
    "Conversation history is context only: never reuse or extrapolate "
    "amounts from it, always call tools for the current question."
)


@dataclass
class AgentResult:
    answer: str
    used_tools: list[str] = field(default_factory=list)


class McpClient(Protocol):
    async def list_tools(self) -> Any: ...

    async def call_tool(
        self, name: str, arguments: dict[str, Any] | None = None
    ) -> Any: ...


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


async def run_agent_stream(
    ollama: AsyncClient,
    mcp: McpClient,
    question: str,
    model: str,
    resident_location_id: int,
    history: list[dict[str, Any]] | None = None,
) -> AsyncGenerator[dict[str, Any], None]:
    tools_result = await mcp.list_tools()
    
    tools = mcp_tools_to_ollama(tools_result.tools)

    system_promt = (
        SYSTEM_PROMT + f"\nThe object (resident_location_id): {resident_location_id}"
    )

    messages: list[dict[str, Any]] = [
        { "role": "system", "content": system_promt },
    ]

    messages.extend(history or [])
    messages.append({
        "role": "user",
        "content": question,
    })

    used: list[str] = []

    for _ in range(MAX_STEP):
        tool_calls: list[Any] = []
        text = ""

        stream = await ollama.chat(
            model=model,
            messages=messages,
            tools=tools,
            options={ "temperature": 0.0 },
            stream=True,
        )

        async for chunk in stream:
            msg = chunk.message

            if msg.tool_calls:
                tool_calls.extend(msg.tool_calls)

                continue

            if msg.content:
                text = text + msg.content

                if not tool_calls:
                    yield {
                        "event": "token",
                        "data": {
                            "text": msg.content
                        }
                    }

        if not tool_calls:
            yield {
                "event": "done",
                "data": {
                    "answer": text,
                    "used_tools": used,
                }
            }

            return

        messages.append(
            {
                "role": "assistant",
                "content": text,
                "tool_calls": tool_calls,
            }
        )

        for call in tool_calls:
            name = call.function.name

            used.append(name)

            yield {
                "event": "tool_start",
                "data": {
                    "name": name,
                },
            }

            try:
                arguments = dict(call.function.arguments or {})

                content = tool_result_to_text(
                    await mcp.call_tool(
                        name,
                        arguments=arguments,
                    )
                )
            except Exception as exc:
                content = f"Error calling tool {name}: {exc}"

            messages.append({
                "role": "tool",
                "tool_name": name,
                "content": content, 
            })

        yield {
            "event": "done",
            "data": {
                "answer": "Not ready to get answer for MAX_STEP",
                "used_tools": used,
            }
        }





async def run_agent(
        ollama: AsyncClient,
        mcp: McpClient,
        question: str,
        model: str,
        resident_location_id: int,
        history: list[dict] | None = None
) -> AgentResult:
    tools_result = await mcp.list_tools()

    tools = mcp_tools_to_ollama(tools_result.tools)

    system_promt = (
        SYSTEM_PROMT + f"\nThe object (resident_location_id): {resident_location_id}"
    )

    messages: list[dict[str, Any]] = [
        { "role": "system", "content": system_promt },
    ]

    messages.extend(history or [])
    messages.append({
        "role": "user",
        "content": question,
    })

    used: list[str] = []

    retried = False

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
            answer = msg.content or ""

            # A model can copy numbers from history instead of calling a tool
            if not used and not retried and re.search(r"\d", answer):
                retried = True
                messages.append({ "role": "assistant", "content": answer })
                messages.append({
                    "role": "user",
                    "content": (
                        "You answered with numbers without calling any tool. "
                        "Call get_monthly_totals for the asked period and "
                        "answer only from its result."
                    ),
                })
                continue

            return AgentResult(
                answer=answer,
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
