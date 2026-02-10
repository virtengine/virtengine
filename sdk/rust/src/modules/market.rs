use tonic::transport::Channel;

use crate::proto::virtengine::market::v1beta5::{
    query_client::QueryClient, QueryLeaseRequest, QueryLeasesRequest, QueryOrderRequest,
    QueryOrdersRequest, QueryParamsRequest, QueryLeaseResponse, QueryLeasesResponse,
    QueryOrderResponse, QueryOrdersResponse, QueryParamsResponse,
};
use crate::{Error, Result};

pub struct MarketModule {
    client: QueryClient<Channel>,
}

impl MarketModule {
    pub fn new(channel: Channel) -> Self {
        Self {
            client: QueryClient::new(channel),
        }
    }

    pub async fn orders(&mut self, state: &str) -> Result<QueryOrdersResponse> {
        let request = QueryOrdersRequest {
            state: state.to_string(),
        };
        self.client
            .orders(request)
            .await
            .map(|res| res.into_inner())
            .map_err(|e| Error::Query(e.to_string()))
    }

    pub async fn order(&mut self, order_id: &str) -> Result<QueryOrderResponse> {
        let request = QueryOrderRequest {
            order_id: order_id.to_string(),
        };
        self.client
            .order(request)
            .await
            .map(|res| res.into_inner())
            .map_err(|e| Error::Query(e.to_string()))
    }

    pub async fn leases(&mut self) -> Result<QueryLeasesResponse> {
        let request = QueryLeasesRequest {};
        self.client
            .leases(request)
            .await
            .map(|res| res.into_inner())
            .map_err(|e| Error::Query(e.to_string()))
    }

    pub async fn lease(&mut self, lease_id: &str) -> Result<QueryLeaseResponse> {
        let request = QueryLeaseRequest {
            lease_id: lease_id.to_string(),
        };
        self.client
            .lease(request)
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
