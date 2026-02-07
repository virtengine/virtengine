"""Wallet utilities for signing transactions."""

from __future__ import annotations

from dataclasses import dataclass
from typing import Tuple
import hashlib

from bip_utils import Bip39SeedGenerator, Bip44, Bip44Coins
from ecdsa import SECP256k1, SigningKey

from virtengine.crypto import bech32_encode_bytes


@dataclass(frozen=True)
class Wallet:
    """VirtEngine wallet for signing transactions."""

    _signing_key: SigningKey
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
        signing_key = SigningKey.from_string(private_key, curve=SECP256k1)
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
        return self._signing_key.get_verifying_key().to_string("compressed")

    def sign(self, message: bytes) -> bytes:
        """Sign a message with the private key."""
        return self._signing_key.sign_deterministic(message, hashfunc=hashlib.sha256)
