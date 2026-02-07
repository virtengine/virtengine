from virtengine.client import VirtEngineClient


def test_default_ws_endpoint_from_http():
    client = VirtEngineClient(rest_endpoint="http://localhost:1317")
    assert client.rpc_ws_endpoint == "ws://localhost:1317/websocket"


def test_default_ws_endpoint_from_https():
    client = VirtEngineClient(rest_endpoint="https://example.com")
    assert client.rpc_ws_endpoint == "wss://example.com/websocket"
