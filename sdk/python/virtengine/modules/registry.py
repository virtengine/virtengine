"""Dynamic module registry for query clients."""

from __future__ import annotations

import importlib
import os
from dataclasses import dataclass
from typing import Dict, Iterable, Optional, Type

import grpc


@dataclass
class ModuleClient:
    name: str
    query_stub: object


class ModuleRegistry:
    """Discover and load query stubs from generated protobuf modules."""

    def __init__(self, channel: grpc.aio.Channel, root_package: str = "virtengine"):
        self._channel = channel
        self._root = root_package
        self._modules: Dict[str, ModuleClient] = {}

    def discover(self) -> None:
        package = importlib.import_module(self._root)
        root_path = os.path.dirname(package.__file__)
        for dirpath, _, filenames in os.walk(root_path):
            for filename in filenames:
                if filename != "query_pb2_grpc.py":
                    continue
                rel_path = os.path.relpath(dirpath, root_path)
                if rel_path == ".":
                    continue
                module_path = ".".join([self._root] + rel_path.split(os.sep) + ["query_pb2_grpc"])
                name = rel_path.split(os.sep)[0]
                if name in self._modules:
                    continue
                stub = self._load_stub(module_path)
                if stub is None:
                    continue
                self._modules[name] = ModuleClient(name=name, query_stub=stub(self._channel))

    def _load_stub(self, module_path: str) -> Optional[Type]:
        try:
            module = importlib.import_module(module_path)
        except ModuleNotFoundError:
            return None
        return getattr(module, "QueryStub", None)

    def get(self, name: str) -> ModuleClient:
        if not self._modules:
            self.discover()
        return self._modules[name]

    def list(self) -> Iterable[ModuleClient]:
        if not self._modules:
            self.discover()
        return self._modules.values()
