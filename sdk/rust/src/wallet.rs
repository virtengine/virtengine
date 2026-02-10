use bech32::{self, ToBase32, Variant};
use bip32::{Mnemonic, XPrv};
use k256::ecdsa::{signature::Signer, Signature, SigningKey};
use ripemd::Ripemd160;
use sha2::{Digest, Sha256};

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
        let path = "m/44'/118'/0'/0/0";
        let xprv = XPrv::derive_from_path(&seed, &path.parse().map_err(|e| Error::Wallet(e.to_string()))?)
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
        let pubkey_bytes = self.signing_key.verifying_key().to_sec1_bytes();
        let sha_hash = Sha256::digest(&pubkey_bytes);
        let ripemd_hash = Ripemd160::digest(&sha_hash);
        bech32::encode(&self.prefix, ripemd_hash.to_base32(), Variant::Bech32)
            .expect("bech32 encoding failed")
    }

    pub fn public_key(&self) -> Vec<u8> {
        self.signing_key.verifying_key().to_sec1_bytes().to_vec()
    }

    pub fn sign(&self, message: &[u8]) -> Vec<u8> {
        let signature: Signature = self.signing_key.sign(message);
        signature.to_vec()
    }
}
