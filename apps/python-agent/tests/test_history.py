from app.api.app import HistoryMessage, trim_history


def test_trim_keeps_last_pair():
    h = [HistoryMessage(role="user", content="a"), HistoryMessage(role="assistant", content="b")]
    assert trim_history(h) == [{"role": "user", "content": "a"},
                               {"role": "assistant", "content": "b"}]


def test_trim_drops_leading_assistant():
    h = [HistoryMessage(role="assistant", content="x"), HistoryMessage(role="user", content="y")]
    assert trim_history(h)[0]["role"] == "user"


def test_trim_budget():
    h = [HistoryMessage(role="user", content="x" * 5000),
         HistoryMessage(role="assistant", content="y" * 5000)]
    assert trim_history(h) == []
