"""Wallet utilities for signing transactions."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Tuple
import hashlib

from bip_utils import Bip39SeedGenerator, Bip44, Bip44Coins
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.primitives.asymmetric import ec
from cryptography.hazmat.primitives.asymmetric.utils import decode_dss_signature
from cryptography.hazmat.primitives.serialization import Encoding, PublicFormat

from virtengine.crypto import bech32_encode_bytes

_SECP256K1_ORDER = int(
    "FFFFFFFFFFFFFFFFFFFFFFFFFFFFFFFEBAAEDCE6AF48A03BBFD25E8CD0364141",
    16,
)
_SECP256K1_HALF_ORDER = _SECP256K1_ORDER // 2


@dataclass(frozen=True)
class Wallet:
    """VirtEngine wallet for signing transactions."""

    _signing_key: ec.EllipticCurvePrivateKey
    prefix: str = "ve"

    @classmethod
    def from_mnemonic(
        cls,
        mnemonic: str,
        account: int = 0,
        index: int = 0,
        prefix: str = "ve",
    ) -> "Wallet":
        """Derive wallet from BIP39 mnemonic."""
        seed = Bip39SeedGenerator(mnemonic).Generate()
        bip44 = Bip44.FromSeed(seed, Bip44Coins.COSMOS)
        account_node = bip44.Purpose().Coin().Account(account).Change(0).AddressIndex(index)
        private_key = account_node.PrivateKey().Raw().ToBytes()
        private_key_int = int.from_bytes(private_key, "big")
        signing_key = ec.derive_private_key(private_key_int, ec.SECP256K1())
        return cls(signing_key, prefix)

    @classmethod
    def generate(cls, prefix: str = "ve") -> Tuple["Wallet", str]:
        """Generate a new wallet with mnemonic."""
        from bip_utils import Bip39MnemonicGenerator, Bip39WordsNum

        mnemonic = Bip39MnemonicGenerator().FromWordsNumber(Bip39WordsNum.WORDS_NUM_24)
        wallet = cls.from_mnemonic(str(mnemonic), prefix=prefix)
        return wallet, str(mnemonic)

    @property
    def address(self) -> str:
        """Return bech32 address."""
        pubkey_bytes = self.public_key
        sha256_hash = hashlib.sha256(pubkey_bytes).digest()
        ripemd160_hash = hashlib.new("ripemd160", sha256_hash).digest()
        return bech32_encode_bytes(self.prefix, ripemd160_hash)

    @property
    def public_key(self) -> bytes:
        """Return compressed public key."""
        return self._signing_key.public_key().public_bytes(
            Encoding.X962,
            PublicFormat.CompressedPoint,
        )

    def sign(self, message: bytes) -> bytes:
        """Sign a message with the private key."""
        signature = self._signing_key.sign(message, ec.ECDSA(hashes.SHA256()))
        r, s = decode_dss_signature(signature)
        if s > _SECP256K1_HALF_ORDER:
            s = _SECP256K1_ORDER - s
        return r.to_bytes(32, "big") + s.to_bytes(32, "big")
