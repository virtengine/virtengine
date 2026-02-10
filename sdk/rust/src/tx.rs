use prost::Message;
use tonic::transport::Channel;

use crate::proto::cosmos::crypto::secp256k1::PubKey;
use crate::proto::cosmos::tx::signing::v1beta1::SignMode;
use crate::proto::cosmos::tx::v1beta1::{
    service_client::ServiceClient, AuthInfo, BroadcastMode, BroadcastTxRequest,
    BroadcastTxResponse, Fee, ModeInfo, SignDoc, SignerInfo, TxBody, TxRaw,
};
use crate::{Error, Result, Wallet};
use prost::bytes::Bytes;

pub fn pack_any<M: Message>(msg: &M, type_url: &str) -> Result<pbjson_types::Any> {
    let mut buf = Vec::new();
    msg.encode(&mut buf)
        .map_err(|e| Error::Tx(e.to_string()))?;
    Ok(pbjson_types::Any {
        type_url: type_url.to_string(),
        value: buf,
    })
}

pub struct TxBuilder {
    chain_id: String,
    gas_limit: u64,
}

impl TxBuilder {
    pub fn new(chain_id: String) -> Self {
        Self {
            chain_id,
            gas_limit: 200_000,
        }
    }

    pub fn build_and_sign(
        &self,
        messages: Vec<pbjson_types::Any>,
        wallet: &Wallet,
        account_number: u64,
        sequence: u64,
        memo: &str,
        fee: Option<Fee>,
    ) -> Result<TxRaw> {
        if messages.is_empty() {
            return Err(Error::Tx("at least one message is required".to_string()));
        }

        let tx_body = TxBody {
            messages,
            memo: memo.to_string(),
            timeout_height: 0,
            unordered: false,
            timeout_timestamp: None,
            extension_options: Vec::new(),
            non_critical_extension_options: Vec::new(),
            memo_extensions: Vec::new(),
        };

        let pubkey = PubKey {
            key: wallet.public_key(),
        };
        let pubkey_any = pack_any(&pubkey, "/cosmos.crypto.secp256k1.PubKey")?;

        let signer_info = SignerInfo {
            public_key: Some(pubkey_any),
            mode_info: Some(ModeInfo {
                sum: Some(crate::proto::cosmos::tx::v1beta1::mode_info::Sum::Single(
                    crate::proto::cosmos::tx::v1beta1::mode_info::Single {
                        mode: SignMode::Direct as i32,
                    },
                )),
            }),
            sequence,
        };

        let fee = fee.unwrap_or(Fee {
            amount: Vec::new(),
            gas_limit: self.gas_limit,
            payer: String::new(),
            granter: String::new(),
        });

        let auth_info = AuthInfo {
            signer_infos: vec![signer_info],
            fee: Some(fee),
            tip: None,
        };

        let body_bytes = tx_body.encode_to_vec();
        let auth_info_bytes = auth_info.encode_to_vec();
        let sign_doc = SignDoc {
            body_bytes: Bytes::from(body_bytes.clone()),
            auth_info_bytes: Bytes::from(auth_info_bytes.clone()),
            chain_id: self.chain_id.clone(),
            account_number,
        };
        let sign_doc_bytes = sign_doc.encode_to_vec();
        let signature = wallet.sign(&sign_doc_bytes);

        Ok(TxRaw {
            body_bytes: Bytes::from(body_bytes),
            auth_info_bytes: Bytes::from(auth_info_bytes),
            signatures: vec![Bytes::from(signature)],
        })
    }
}

pub struct TxBroadcaster {
    client: ServiceClient<Channel>,
}

impl TxBroadcaster {
    pub fn new(channel: Channel) -> Self {
        Self {
            client: ServiceClient::new(channel),
        }
    }

    pub async fn broadcast(
        &mut self,
        tx: TxRaw,
        mode: BroadcastMode,
    ) -> Result<BroadcastTxResponse> {
        let request = BroadcastTxRequest {
            tx_bytes: tx.encode_to_vec(),
            mode: mode as i32,
        };
        self.client
            .broadcast_tx(request)
            .await
            .map(|res| res.into_inner())
            .map_err(|e| Error::Tx(e.to_string()))
    }
}
