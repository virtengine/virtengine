"""Cryptography helpers for VirtEngine."""

from __future__ import annotations

from typing import Iterable

from bech32 import bech32_encode, convertbits


class Bech32EncodingError(ValueError):
    """Raised when bech32 encoding fails."""


def bech32_encode_bytes(hrp: str, data: bytes) -> str:
    """Encode raw bytes into bech32 string with provided HRP."""
    if not hrp:
        raise Bech32EncodingError("HRP must be non-empty")
    if not data:
        raise Bech32EncodingError("Data must be non-empty")
    five_bit = convertbits(data, 8, 5, True)
    if five_bit is None:
        raise Bech32EncodingError("Failed to convert data to 5-bit words")
    return bech32_encode(hrp, list(five_bit))


def bech32_encode_values(hrp: str, values: Iterable[int]) -> str:
    """Encode 5-bit values into bech32 string."""
    value_list = list(values)
    if not value_list:
        raise Bech32EncodingError("Values must be non-empty")
    return bech32_encode(hrp, value_list)
