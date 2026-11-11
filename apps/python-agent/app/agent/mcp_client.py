import sys
import asyncio

from pathlib import Path

from mcp import Client, StdioServerParameters

PROJECT_ROOT = Path(__file__).resolve().parents[2]


def build_server_params() -> StdioServerParameters:
    """Parameters running MCP-server as subprocess via stdio"""
    return StdioServerParameters(
        command=sys.executable,
        args=["-m", "app.mcp_server"],
        cwd=str(PROJECT_ROOT),
    )

async def main() -> None:
    async with Client(build_server_params()) as client:
        tools = await client.list_tools()
        for t in tools.tools:
            print(t.name)

        result = await client.call_tool(
            "get_monthly_totals",
            {"resident_location_id": 1},
        )
        print(result.is_error)
        if result.content:
            for block in result.content:
                print(block.text)
        else:
            print(result.structured_content)


if __name__ == "__main__":
    asyncio.run(main())