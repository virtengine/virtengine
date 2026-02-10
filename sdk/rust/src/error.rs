use thiserror::Error;

#[derive(Debug, Error)]
pub enum Error {
    #[error("connection error: {0}")]
    Connection(String),
    #[error("wallet error: {0}")]
    Wallet(String),
    #[error("query error: {0}")]
    Query(String),
    #[error("transaction error: {0}")]
    Tx(String),
    #[error("missing wallet")]
    NoWallet,
}
