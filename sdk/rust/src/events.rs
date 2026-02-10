use futures_util::{stream::BoxStream, StreamExt};
use serde_json::Value;
use tokio_tungstenite::{connect_async, tungstenite::Message};

use crate::{Error, Result};

pub struct EventSubscriber {
    ws_endpoint: String,
}

impl EventSubscriber {
    pub fn new(ws_endpoint: &str) -> Self {
        Self {
            ws_endpoint: ws_endpoint.to_string(),
        }
    }

    pub async fn subscribe(&self, query: &str) -> Result<BoxStream<'static, Result<Value>>> {
        let (ws_stream, _) = connect_async(&self.ws_endpoint)
            .await
            .map_err(|e| Error::Connection(e.to_string()))?;
        let (mut write, read) = ws_stream.split();
        let request = serde_json::json!({
            "jsonrpc": "2.0",
            "id": "1",
            "method": "subscribe",
            "params": { "query": query },
        });
        write
            .send(Message::Text(request.to_string()))
            .await
            .map_err(|e| Error::Connection(e.to_string()))?;

        let stream = read.filter_map(|msg| async move {
            match msg {
                Ok(Message::Text(text)) => {
                    let value = serde_json::from_str::<Value>(&text)
                        .map_err(|e| Error::Query(e.to_string()));
                    Some(value)
                }
                Ok(Message::Binary(bytes)) => {
                    let value = serde_json::from_slice::<Value>(&bytes)
                        .map_err(|e| Error::Query(e.to_string()));
                    Some(value)
                }
                Ok(_) => None,
                Err(e) => Some(Err(Error::Connection(e.to_string()))),
            }
        });

        Ok(Box::pin(stream))
    }
}
