from typing import Any

from mcp.types import Tool


def mcp_tools_to_ollama(tools: list[Tool]) -> list[dict[str, Any]]:
    """MCP-tool -> list of tools in format of Ollama chat(tools=...)"""
    return [
        {
            "type": "function",
            "function": {
                "name": tool.name,
                "description": tool.description,
                "parameters": tool.input_schema or { "type": "object", "properties": {} },
            },
        }
        for tool in tools
    ]