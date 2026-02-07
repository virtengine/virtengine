use std::sync::Arc;

use prost::Message;
use tonic::transport::Channel;

use crate::modules::{EscrowModule, MarketModule, VEIDModule};
use crate::proto::cosmos::auth::v1beta1::{
    query_client::QueryClient as AuthQueryClient, BaseAccount, ModuleAccount, QueryAccountRequest,
};
use crate::tx::{TxBroadcaster, TxBuilder};
use crate::{Error, EventSubscriber, Result, Wallet};

pub struct AccountInfo {
    pub account_number: u64,
    pub sequence: u64,
    pub address: String,
}

pub struct VirtEngineClient {
    channel: Channel,
    chain_id: String,
    wallet: Option<Arc<Wallet>>,

    pub veid: VEIDModule,
    pub market: MarketModule,
    pub escrow: EscrowModule,

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

    pub async fn from_mnemonic(endpoint: &str, chain_id: &str, mnemonic: &str) -> Result<Self> {
        let wallet = Wallet::from_mnemonic(mnemonic, "ve")?;
        let client = Self::connect(endpoint, chain_id).await?;
        Ok(client.with_wallet(wallet))
    }

    pub async fn account(&self, address: &str) -> Result<AccountInfo> {
        let mut client = AuthQueryClient::new(self.channel.clone());
        let request = QueryAccountRequest {
            address: address.to_string(),
        };
        let response = client
            .account(request)
            .await
            .map_err(|e| Error::Query(e.to_string()))?
            .into_inner();
        let account_any = response
            .account
            .ok_or_else(|| Error::Query("missing account".to_string()))?;
        let base_account = Self::unpack_account(account_any)?;
        Ok(AccountInfo {
            account_number: base_account.account_number,
            sequence: base_account.sequence,
            address: base_account.address,
        })
    }

    pub async fn sign_and_broadcast(
        &mut self,
        messages: Vec<pbjson_types::Any>,
        memo: &str,
        fee: Option<crate::proto::cosmos::tx::v1beta1::Fee>,
        mode: crate::proto::cosmos::tx::v1beta1::BroadcastMode,
    ) -> Result<crate::proto::cosmos::tx::v1beta1::BroadcastTxResponse> {
        let wallet = self.wallet.as_ref().ok_or(Error::NoWallet)?;
        let account = self.account(&wallet.address()).await?;
        let tx = self.tx_builder.build_and_sign(
            messages,
            wallet.as_ref(),
            account.account_number,
            account.sequence,
            memo,
            fee,
        )?;
        self.tx_broadcaster.broadcast(tx, mode).await
    }

    pub fn events(&self, ws_endpoint: &str) -> EventSubscriber {
        EventSubscriber::new(ws_endpoint)
    }

    fn unpack_account(account_any: pbjson_types::Any) -> Result<BaseAccount> {
        match account_any.type_url.as_str() {
            "/cosmos.auth.v1beta1.BaseAccount" => {
                BaseAccount::decode(account_any.value.as_slice())
                    .map_err(|e| Error::Query(e.to_string()))
            }
            "/cosmos.auth.v1beta1.ModuleAccount" => {
                let module = ModuleAccount::decode(account_any.value.as_slice())
                    .map_err(|e| Error::Query(e.to_string()))?;
                module
                    .base_account
                    .ok_or_else(|| Error::Query("missing base account".to_string()))
            }
            _ => Err(Error::Query("unsupported account type".to_string())),
        }
    }
}
