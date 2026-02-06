# Task 31K: Python and Rust SDK Clients

**vibe-kanban ID:** `5e5df72b-dd28-4d3a-b2a8-91e50a21f72b`

## Overview

| Field | Value |
|-------|-------|
| **ID** | 31K |
| **Title** | feat(sdk): Python and Rust SDK clients |
| **Priority** | P2 |
| **Wave** | 4 |
| **Estimated LOC** | 6000 |
| **Duration** | 4-5 weeks |
| **Dependencies** | TypeScript SDK exists as reference |
| **Blocking** | None |

---

## Problem Statement

The TypeScript SDK exists in `sdk/ts/`, but developers in other ecosystems need native SDK:
- Python for data scientists and ML engineers
- Rust for high-performance applications and smart contracts
- Both for broader developer adoption

Each SDK should provide:
- Transaction signing and broadcasting
- Query APIs for all modules
- Subscription to chain events
- Type-safe interfaces

### Current State Analysis

```
sdk/ts/                         ✅ TypeScript SDK exists
sdk/python/                     ❌ Does not exist
sdk/rust/                       ❌ Does not exist (rs/ is smart contracts)
protobuf definitions:           ✅ Available for codegen
```

---

## Acceptance Criteria

### AC-1: Python SDK
- [ ] Package structure with Poetry/pip
- [ ] Transaction building and signing
- [ ] All module query clients
- [ ] WebSocket event subscription
- [ ] Type hints (PEP 484)
- [ ] PyPI publishable

### AC-2: Rust SDK
- [ ] Crate structure with Cargo
- [ ] Transaction building and signing
- [ ] All module query clients
- [ ] Async/await support (tokio)
- [ ] crates.io publishable

### AC-3: Documentation
- [ ] API reference (auto-generated)
- [ ] Getting started guide
- [ ] Example applications
- [ ] Migration guide from raw gRPC

### AC-4: Testing
- [ ] Unit tests (>80% coverage)
- [ ] Integration tests with testnet
- [ ] CI/CD for publishing

---

## Technical Requirements

### Python SDK

```python
# sdk/python/virtengine/__init__.py

from virtengine.client import VirtEngineClient
from virtengine.wallet import Wallet
from virtengine.types import *

__version__ = "0.1.0"
__all__ = ["VirtEngineClient", "Wallet"]
```

```python
# sdk/python/virtengine/client.py

from typing import Optional, List
import grpc
from google.protobuf import any_pb2

from virtengine.modules.veid import VEIDModule
from virtengine.modules.market import MarketModule
from virtengine.modules.escrow import EscrowModule
from virtengine.wallet import Wallet
from virtengine.tx import TxBuilder, TxBroadcaster


class VirtEngineClient:
    """Main client for interacting with VirtEngine blockchain."""
    
    def __init__(
        self,
        grpc_endpoint: str = "localhost:9090",
        rest_endpoint: str = "http://localhost:1317",
        chain_id: str = "virtengine-1",
        wallet: Optional[Wallet] = None,
    ):
        self.grpc_endpoint = grpc_endpoint
        self.rest_endpoint = rest_endpoint
        self.chain_id = chain_id
        self.wallet = wallet
        
        # Initialize gRPC channel
        self._channel = grpc.aio.insecure_channel(grpc_endpoint)
        
        # Initialize modules
        self.veid = VEIDModule(self._channel)
        self.market = MarketModule(self._channel)
        self.escrow = EscrowModule(self._channel)
        
        # Transaction handling
        self._tx_builder = TxBuilder(chain_id)
        self._tx_broadcaster = TxBroadcaster(self._channel)
    
    @classmethod
    def from_mnemonic(
        cls,
        mnemonic: str,
        grpc_endpoint: str = "localhost:9090",
        **kwargs
    ) -> "VirtEngineClient":
        """Create client with wallet from mnemonic."""
        wallet = Wallet.from_mnemonic(mnemonic)
        return cls(grpc_endpoint=grpc_endpoint, wallet=wallet, **kwargs)
    
    async def sign_and_broadcast(
        self,
        messages: List[any_pb2.Any],
        memo: str = "",
        fee: Optional[dict] = None,
    ) -> "TxResponse":
        """Sign and broadcast a transaction with given messages."""
        if not self.wallet:
            raise ValueError("Wallet required for signing")
        
        # Get account info for sequence
        account = await self._get_account()
        
        # Build transaction
        tx = self._tx_builder.build(
            messages=messages,
            signer=self.wallet,
            account_number=account.account_number,
            sequence=account.sequence,
            memo=memo,
            fee=fee,
        )
        
        # Broadcast
        return await self._tx_broadcaster.broadcast(tx)
    
    async def close(self):
        """Close the gRPC channel."""
        await self._channel.close()
    
    async def __aenter__(self):
        return self
    
    async def __aexit__(self, exc_type, exc_val, exc_tb):
        await self.close()
```

```python
# sdk/python/virtengine/modules/veid.py

from typing import Optional
import grpc

from virtengine.proto.veid.v1 import query_pb2, query_pb2_grpc
from virtengine.proto.veid.v1 import tx_pb2
from virtengine.types import IdentityRecord, Scope, ScopeType


class VEIDModule:
    """VEID module queries and transactions."""
    
    def __init__(self, channel: grpc.aio.Channel):
        self._query_client = query_pb2_grpc.QueryStub(channel)
    
    async def get_identity(self, address: str) -> Optional[IdentityRecord]:
        """Get identity record for an address."""
        try:
            request = query_pb2.QueryIdentityRequest(address=address)
            response = await self._query_client.Identity(request)
            return IdentityRecord.from_proto(response.identity)
        except grpc.RpcError as e:
            if e.code() == grpc.StatusCode.NOT_FOUND:
                return None
            raise
    
    async def get_scope(self, scope_id: str) -> Optional[Scope]:
        """Get a specific scope by ID."""
        request = query_pb2.QueryScopeRequest(scope_id=scope_id)
        response = await self._query_client.Scope(request)
        return Scope.from_proto(response.scope)
    
    async def list_scopes(
        self,
        address: str,
        scope_type: Optional[ScopeType] = None,
    ) -> list[Scope]:
        """List scopes for an address."""
        request = query_pb2.QueryScopesRequest(
            address=address,
            scope_type=scope_type.value if scope_type else None,
        )
        response = await self._query_client.Scopes(request)
        return [Scope.from_proto(s) for s in response.scopes]
    
    def submit_scope(
        self,
        scope_type: ScopeType,
        encrypted_data: bytes,
        client_sig: bytes,
        envelope_data: dict,
    ) -> tx_pb2.MsgSubmitScope:
        """Build a MsgSubmitScope (requires signing separately)."""
        return tx_pb2.MsgSubmitScope(
            scope_type=scope_type.value,
            encrypted_data=encrypted_data,
            client_sig=client_sig,
            envelope=envelope_data,
        )
```

```python
# sdk/python/virtengine/wallet.py

from typing import Tuple
from bip_utils import Bip39SeedGenerator, Bip44, Bip44Coins
from ecdsa import SigningKey, SECP256k1
import hashlib

from virtengine.crypto import bech32_encode


class Wallet:
    """VirtEngine wallet for signing transactions."""
    
    def __init__(self, private_key: bytes, prefix: str = "ve"):
        self._private_key = SigningKey.from_string(private_key, curve=SECP256k1)
        self._public_key = self._private_key.get_verifying_key()
        self._prefix = prefix
    
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
        return cls(private_key, prefix)
    
    @classmethod
    def generate(cls, prefix: str = "ve") -> Tuple["Wallet", str]:
        """Generate new wallet with mnemonic."""
        from bip_utils import Bip39MnemonicGenerator, Bip39WordsNum
        
        mnemonic = Bip39MnemonicGenerator().FromWordsNumber(Bip39WordsNum.WORDS_NUM_24)
        wallet = cls.from_mnemonic(str(mnemonic), prefix=prefix)
        return wallet, str(mnemonic)
    
    @property
    def address(self) -> str:
        """Get bech32 address."""
        pubkey_bytes = self._public_key.to_string("compressed")
        sha256_hash = hashlib.sha256(pubkey_bytes).digest()
        ripemd160 = hashlib.new("ripemd160", sha256_hash).digest()
        return bech32_encode(self._prefix, ripemd160)
    
    @property
    def public_key(self) -> bytes:
        """Get compressed public key."""
        return self._public_key.to_string("compressed")
    
    def sign(self, message: bytes) -> bytes:
        """Sign a message with private key."""
        return self._private_key.sign_deterministic(
            message,
            hashfunc=hashlib.sha256,
        )
```

### Rust SDK

```rust
// sdk/rust/src/lib.rs

pub mod client;
pub mod modules;
pub mod wallet;
pub mod tx;
pub mod types;
pub mod error;

pub use client::VirtEngineClient;
pub use wallet::Wallet;
pub use error::Error;

pub type Result<T> = std::result::Result<T, Error>;
```

```rust
// sdk/rust/src/client.rs

use tonic::transport::Channel;
use std::sync::Arc;

use crate::modules::{VEIDModule, MarketModule, EscrowModule};
use crate::wallet::Wallet;
use crate::tx::{TxBuilder, TxBroadcaster};
use crate::{Error, Result};

pub struct VirtEngineClient {
    channel: Channel,
    chain_id: String,
    wallet: Option<Arc<Wallet>>,
    
    // Module clients
    pub veid: VEIDModule,
    pub market: MarketModule,
    pub escrow: EscrowModule,
    
    // Tx handling
    tx_builder: TxBuilder,
    tx_broadcaster: TxBroadcaster,
}

impl VirtEngineClient {
    pub async fn connect(endpoint: &str, chain_id: &str) -> Result<Self> {
        let channel = Channel::from_shared(endpoint.to_string())
            .map_err(|e| Error::Connection(e.to_string()))?
            .connect()
            .await
            .map_err(|e| Error::Connection(e.to_string()))?;
        
        Ok(Self {
            veid: VEIDModule::new(channel.clone()),
            market: MarketModule::new(channel.clone()),
            escrow: EscrowModule::new(channel.clone()),
            tx_builder: TxBuilder::new(chain_id.to_string()),
            tx_broadcaster: TxBroadcaster::new(channel.clone()),
            channel,
            chain_id: chain_id.to_string(),
            wallet: None,
        })
    }
    
    pub fn with_wallet(mut self, wallet: Wallet) -> Self {
        self.wallet = Some(Arc::new(wallet));
        self
    }
    
    pub async fn from_mnemonic(
        endpoint: &str,
        chain_id: &str,
        mnemonic: &str,
    ) -> Result<Self> {
        let wallet = Wallet::from_mnemonic(mnemonic, "ve")?;
        let client = Self::connect(endpoint, chain_id).await?;
        Ok(client.with_wallet(wallet))
    }
    
    pub async fn sign_and_broadcast<M: prost::Message>(
        &self,
        messages: Vec<M>,
        memo: &str,
        fee: Option<Fee>,
    ) -> Result<TxResponse> {
        let wallet = self.wallet.as_ref()
            .ok_or(Error::NoWallet)?;
        
        // Get account info
        let account = self.get_account(wallet.address()).await?;
        
        // Build and sign transaction
        let tx = self.tx_builder.build_and_sign(
            messages,
            wallet.as_ref(),
            account.account_number,
            account.sequence,
            memo,
            fee,
        )?;
        
        // Broadcast
        self.tx_broadcaster.broadcast(tx).await
    }
}
```

```rust
// sdk/rust/src/modules/veid.rs

use tonic::transport::Channel;

use crate::types::{IdentityRecord, Scope, ScopeType};
use crate::proto::veid::v1::{
    query_client::QueryClient,
    QueryIdentityRequest,
    QueryScopeRequest,
    QueryScopesRequest,
};
use crate::{Error, Result};

pub struct VEIDModule {
    client: QueryClient<Channel>,
}

impl VEIDModule {
    pub fn new(channel: Channel) -> Self {
        Self {
            client: QueryClient::new(channel),
        }
    }
    
    pub async fn get_identity(&mut self, address: &str) -> Result<Option<IdentityRecord>> {
        let request = QueryIdentityRequest {
            address: address.to_string(),
        };
        
        match self.client.identity(request).await {
            Ok(response) => {
                let identity = response.into_inner().identity
                    .map(IdentityRecord::from_proto)
                    .transpose()?;
                Ok(identity)
            }
            Err(status) if status.code() == tonic::Code::NotFound => Ok(None),
            Err(e) => Err(Error::Query(e.to_string())),
        }
    }
    
    pub async fn get_scope(&mut self, scope_id: &str) -> Result<Option<Scope>> {
        let request = QueryScopeRequest {
            scope_id: scope_id.to_string(),
        };
        
        let response = self.client.scope(request).await
            .map_err(|e| Error::Query(e.to_string()))?;
        
        let scope = response.into_inner().scope
            .map(Scope::from_proto)
            .transpose()?;
        
        Ok(scope)
    }
    
    pub async fn list_scopes(
        &mut self,
        address: &str,
        scope_type: Option<ScopeType>,
    ) -> Result<Vec<Scope>> {
        let request = QueryScopesRequest {
            address: address.to_string(),
            scope_type: scope_type.map(|t| t as i32).unwrap_or(0),
        };
        
        let response = self.client.scopes(request).await
            .map_err(|e| Error::Query(e.to_string()))?;
        
        response.into_inner().scopes
            .into_iter()
            .map(Scope::from_proto)
            .collect()
    }
}
```

```rust
// sdk/rust/src/wallet.rs

use bip32::{Mnemonic, XPrv};
use k256::ecdsa::{SigningKey, signature::Signer};
use sha2::{Sha256, Digest};
use ripemd::Ripemd160;
use bech32::{self, ToBase32, Variant};

use crate::{Error, Result};

pub struct Wallet {
    signing_key: SigningKey,
    prefix: String,
}

impl Wallet {
    pub fn from_mnemonic(mnemonic: &str, prefix: &str) -> Result<Self> {
        let mnemonic = Mnemonic::new(mnemonic, Default::default())
            .map_err(|e| Error::Wallet(e.to_string()))?;
        
        let seed = mnemonic.to_seed("");
        let path = "m/44'/118'/0'/0/0"; // Cosmos derivation path
        
        let xprv = XPrv::derive_from_path(&seed, &path.parse().unwrap())
            .map_err(|e| Error::Wallet(e.to_string()))?;
        
        let signing_key = SigningKey::from_bytes(&xprv.private_key().to_bytes())
            .map_err(|e| Error::Wallet(e.to_string()))?;
        
        Ok(Self {
            signing_key,
            prefix: prefix.to_string(),
        })
    }
    
    pub fn generate(prefix: &str) -> Result<(Self, String)> {
        let mnemonic = Mnemonic::random(&mut rand::thread_rng(), Default::default());
        let phrase = mnemonic.phrase().to_string();
        let wallet = Self::from_mnemonic(&phrase, prefix)?;
        Ok((wallet, phrase))
    }
    
    pub fn address(&self) -> String {
        let pubkey = self.signing_key.verifying_key();
        let pubkey_bytes = pubkey.to_sec1_bytes();
        
        let sha_hash = Sha256::digest(&pubkey_bytes);
        let ripemd_hash = Ripemd160::digest(&sha_hash);
        
        bech32::encode(&self.prefix, ripemd_hash.to_base32(), Variant::Bech32)
            .expect("bech32 encoding failed")
    }
    
    pub fn public_key(&self) -> Vec<u8> {
        self.signing_key.verifying_key().to_sec1_bytes().to_vec()
    }
    
    pub fn sign(&self, message: &[u8]) -> Vec<u8> {
        use k256::ecdsa::Signature;
        let signature: Signature = self.signing_key.sign(message);
        signature.to_vec()
    }
}
```

### Package Configuration

```toml
# sdk/python/pyproject.toml
[tool.poetry]
name = "virtengine"
version = "0.1.0"
description = "Python SDK for VirtEngine blockchain"
authors = ["VirtEngine <sdk@virtengine.io>"]
license = "Apache-2.0"
readme = "README.md"
homepage = "https://github.com/virtengine/virtengine"
repository = "https://github.com/virtengine/virtengine/tree/main/sdk/python"
documentation = "https://docs.virtengine.io/sdk/python"
keywords = ["blockchain", "cosmos", "sdk", "identity"]
classifiers = [
    "Development Status :: 4 - Beta",
    "Intended Audience :: Developers",
    "Topic :: Software Development :: Libraries",
]

[tool.poetry.dependencies]
python = "^3.10"
grpcio = "^1.60.0"
grpcio-tools = "^1.60.0"
protobuf = "^4.25.0"
bip-utils = "^2.9.0"
ecdsa = "^0.18.0"
bech32 = "^1.2.0"

[tool.poetry.group.dev.dependencies]
pytest = "^7.4.0"
pytest-asyncio = "^0.23.0"
pytest-cov = "^4.1.0"
mypy = "^1.8.0"
ruff = "^0.1.9"

[build-system]
requires = ["poetry-core"]
build-backend = "poetry.core.masonry.api"
```

```toml
# sdk/rust/Cargo.toml
[package]
name = "virtengine-sdk"
version = "0.1.0"
edition = "2021"
authors = ["VirtEngine <sdk@virtengine.io>"]
license = "Apache-2.0"
description = "Rust SDK for VirtEngine blockchain"
repository = "https://github.com/virtengine/virtengine/tree/main/sdk/rust"
documentation = "https://docs.virtengine.io/sdk/rust"
keywords = ["blockchain", "cosmos", "sdk", "identity"]
categories = ["api-bindings", "cryptography"]

[dependencies]
tonic = { version = "0.11", features = ["tls"] }
prost = "0.12"
tokio = { version = "1", features = ["full"] }
k256 = { version = "0.13", features = ["ecdsa"] }
bip32 = "0.5"
sha2 = "0.10"
ripemd = "0.1"
bech32 = "0.9"
rand = "0.8"
thiserror = "1.0"
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"

[dev-dependencies]
tokio-test = "0.4"
wiremock = "0.5"

[build-dependencies]
tonic-build = "0.11"
```

---

## Directory Structure

```
sdk/
├── python/
│   ├── virtengine/
│   │   ├── __init__.py
│   │   ├── client.py
│   │   ├── wallet.py
│   │   ├── tx.py
│   │   ├── crypto.py
│   │   ├── modules/
│   │   │   ├── __init__.py
│   │   │   ├── veid.py
│   │   │   ├── market.py
│   │   │   └── escrow.py
│   │   ├── types/
│   │   │   └── __init__.py
│   │   └── proto/
│   │       └── (generated)
│   ├── tests/
│   │   ├── test_client.py
│   │   ├── test_wallet.py
│   │   └── test_modules.py
│   ├── examples/
│   │   ├── submit_veid.py
│   │   ├── create_order.py
│   │   └── watch_events.py
│   ├── pyproject.toml
│   └── README.md
└── rust/
    ├── src/
    │   ├── lib.rs
    │   ├── client.rs
    │   ├── wallet.rs
    │   ├── tx.rs
    │   ├── error.rs
    │   ├── modules/
    │   │   ├── mod.rs
    │   │   ├── veid.rs
    │   │   ├── market.rs
    │   │   └── escrow.rs
    │   ├── types/
    │   │   └── mod.rs
    │   └── proto/
    │       └── (generated)
    ├── tests/
    │   └── integration.rs
    ├── examples/
    │   ├── submit_veid.rs
    │   └── create_order.rs
    ├── Cargo.toml
    └── README.md
```

---

## Testing Requirements

### Python Tests
```bash
poetry run pytest --cov=virtengine --cov-report=html
```

### Rust Tests
```bash
cargo test
cargo test --features integration
```

### Integration Tests
- Connect to testnet
- Submit real transactions
- Query chain state

---

## CI/CD Publishing

```yaml
# .github/workflows/sdk-publish.yaml
name: Publish SDKs

on:
  release:
    types: [published]

jobs:
  publish-python:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-python@v5
        with:
          python-version: '3.11'
      - name: Install Poetry
        run: pip install poetry
      - name: Build and publish
        working-directory: sdk/python
        run: |
          poetry build
          poetry publish --username __token__ --password ${{ secrets.PYPI_TOKEN }}

  publish-rust:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: dtolnay/rust-toolchain@stable
      - name: Publish to crates.io
        working-directory: sdk/rust
        run: cargo publish --token ${{ secrets.CRATES_TOKEN }}
```
