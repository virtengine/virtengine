//! Rust SDK for VirtEngine.

pub mod client;
pub mod events;
pub mod error;
pub mod modules;
pub mod proto;
pub mod tx;
pub mod types;
pub mod wallet;

pub use client::VirtEngineClient;
pub use events::EventSubscriber;
pub use error::Error;
pub use wallet::Wallet;

pub type Result<T> = std::result::Result<T, Error>;
