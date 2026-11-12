import asyncio
from types import SimpleNamespace

from app.agent import loop


def _reply(content="", tool_calls=None):
    return SimpleNamespace(message=SimpleNamespace(content=content, tool_calls=tool_calls))


def _call(name):
    return SimpleNamespace(function=SimpleNamespace(name=name, arguments={}))


class FakeOllama:
    def __init__(self, replies):
        self.replies = list(replies)
        self.seen = []

    async def chat(self, **kwargs):
        self.seen.append(list(kwargs["messages"]))
        return self.replies.pop(0)


class FakeMcp:
    def __init__(self):
        self.calls = []

    async def list_tools(self):
        return SimpleNamespace(tools=[])

    async def call_tool(self, name, arguments=None):
        self.calls.append(name)
        return SimpleNamespace(structured_content={"total": "372.49"}, content=[])


def test_numeric_answer_without_tool_is_retried(monkeypatch):
    monkeypatch.setattr(loop, "mcp_tools_to_ollama", lambda tools: [])

    ollama = FakeOllama([
        _reply("В апреле 2026 вы потратили 1234.50 BYN."),
        _reply(tool_calls=[_call("get_monthly_totals")]),
        _reply("В апреле 2026 расходы 372.49 BYN."),
    ])
    mcp = FakeMcp()

    result = asyncio.run(loop.run_agent(ollama, mcp, "а в апреле?", "model", 1))

    assert result.used_tools == ["get_monthly_totals"]
    assert "372.49" in result.answer
    assert len(ollama.seen) == 3
    assert ollama.seen[1][-1]["role"] == "user"
    assert "without calling any tool" in ollama.seen[1][-1]["content"]
