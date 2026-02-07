use tonic::transport::Channel;

use crate::proto::virtengine::escrow::v1::{
    query_client::QueryClient, QueryAccountRequest, QueryBalancesRequest, QueryParamsRequest,
    QueryAccountResponse, QueryBalancesResponse, QueryParamsResponse,
};
use crate::{Error, Result};

pub struct EscrowModule {
    client: QueryClient<Channel>,
}

impl EscrowModule {
    pub fn new(channel: Channel) -> Self {
        Self {
            client: QueryClient::new(channel),
        }
    }

    pub async fn balances(&mut self, owner: &str) -> Result<QueryBalancesResponse> {
        let request = QueryBalancesRequest {
            owner: owner.to_string(),
        };
        self.client
            .balances(request)
            .await
            .map(|res| res.into_inner())
            .map_err(|e| Error::Query(e.to_string()))
    }

    pub async fn account(&mut self, address: &str) -> Result<QueryAccountResponse> {
        let request = QueryAccountRequest {
            address: address.to_string(),
        };
        self.client
            .account(request)
            .await
            .map(|res| res.into_inner())
            .map_err(|e| Error::Query(e.to_string()))
    }

    pub async fn params(&mut self) -> Result<QueryParamsResponse> {
        let request = QueryParamsRequest {};
        self.client
            .params(request)
            .await
            .map(|res| res.into_inner())
            .map_err(|e| Error::Query(e.to_string()))
    }
}
