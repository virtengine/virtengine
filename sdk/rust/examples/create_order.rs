use virtengine_sdk::tx::pack_any;
use virtengine_sdk::VirtEngineClient;

use virtengine_sdk::proto::virtengine::market::v1beta5::MsgCreateBid;

#[tokio::main]
async fn main() -> Result<(), Box<dyn std::error::Error>> {
    let mut client = VirtEngineClient::from_mnemonic(
        "http://localhost:9090",
        "virtengine-1",
        "your mnemonic here",
    )
    .await?;

    let msg = MsgCreateBid {
        order: None,
        provider: None,
        price: None,
    };
    let msg_any = pack_any(&msg, "/virtengine.market.v1beta5.MsgCreateBid")?;
    let response = client
        .sign_and_broadcast(vec![msg_any], "", None, virtengine_sdk::proto::cosmos::tx::v1beta1::BroadcastMode::Sync)
        .await?;
    println!("{:?}", response.tx_response);
    Ok(())
}
