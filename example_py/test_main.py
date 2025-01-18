import json
from .main import deserialize, serialize


def test_serialize_and_deserialize():
    d = {
        "HandshakeResult": {
            "IsOk": True,
            "Message": "This is a message",
            "TimeoutMs": 10,
            "ScenarioFreqSecs": 20,
            "NextStartDatetime": "20250505",
        },
    }
    obj = deserialize(d)
    assert obj.HandshakeResult.IsOk
    assert obj.HandshakeResult.Message == "This is a message"

    assert serialize(obj) == d
