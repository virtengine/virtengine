use virtengine_sdk::Wallet;

#[test]
fn wallet_generates_address_prefix() {
    let (wallet, _mnemonic) = Wallet::generate("ve").expect("wallet should generate");
    let address = wallet.address();
    assert!(address.starts_with("ve"));
}
