from virtengine.wallet import Wallet


def test_wallet_generate_and_address_prefix():
    wallet, mnemonic = Wallet.generate(prefix="ve")
    assert mnemonic
    assert wallet.address.startswith("ve")


def test_wallet_signing_roundtrip():
    wallet, _ = Wallet.generate()
    message = b"virtengine-test"
    signature = wallet.sign(message)
    assert isinstance(signature, bytes)
    assert signature
