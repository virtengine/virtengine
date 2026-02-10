"""Module accessors."""

from virtengine.modules.escrow import EscrowModule
from virtengine.modules.market import MarketModule
from virtengine.modules.registry import ModuleClient, ModuleRegistry
from virtengine.modules.veid import VEIDModule

__all__ = [
    "EscrowModule",
    "MarketModule",
    "ModuleClient",
    "ModuleRegistry",
    "VEIDModule",
]
