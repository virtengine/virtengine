import grpc

from virtengine.modules.registry import ModuleRegistry


def test_module_registry_discovers_modules():
    channel = grpc.aio.insecure_channel("localhost:9090")
    registry = ModuleRegistry(channel)
    registry.discover()
    names = {module.name for module in registry.list()}
    assert "veid" in names
    assert "market" in names
    assert "escrow" in names
