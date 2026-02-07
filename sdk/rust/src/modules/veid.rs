use tonic::transport::Channel;

use crate::proto::virtengine::veid::v1::{
    query_client::QueryClient, QueryApprovedClientsRequest, QueryIdentityRecordRequest,
    QueryIdentityRequest, QueryParamsRequest, QueryScopeRequest, QueryScopesByTypeRequest,
    QueryScopesRequest, QueryApprovedClientsResponse, QueryIdentityRecordResponse,
    QueryIdentityResponse, QueryParamsResponse, QueryScopeResponse, QueryScopesByTypeResponse,
    QueryScopesResponse,
};
use crate::{Error, Result};

pub struct VEIDModule {
    client: QueryClient<Channel>,
}

impl VEIDModule {
    pub fn new(channel: Channel) -> Self {
        Self {
            client: QueryClient::new(channel),
        }
    }

    pub async fn identity(&mut self, account_address: &str) -> Result<QueryIdentityResponse> {
        let request = QueryIdentityRequest {
            account_address: account_address.to_string(),
        };
        self.client
            .identity(request)
            .await
            .map(|res| res.into_inner())
            .map_err(|e| Error::Query(e.to_string()))
    }

    pub async fn identity_record(
        &mut self,
        account_address: &str,
    ) -> Result<QueryIdentityRecordResponse> {
        let request = QueryIdentityRecordRequest {
            account_address: account_address.to_string(),
        };
        self.client
            .identity_record(request)
            .await
            .map(|res| res.into_inner())
            .map_err(|e| Error::Query(e.to_string()))
    }

    pub async fn scope(
        &mut self,
        account_address: &str,
        scope_id: &str,
    ) -> Result<QueryScopeResponse> {
        let request = QueryScopeRequest {
            account_address: account_address.to_string(),
            scope_id: scope_id.to_string(),
        };
        self.client
            .scope(request)
            .await
            .map(|res| res.into_inner())
            .map_err(|e| Error::Query(e.to_string()))
    }

    pub async fn scopes(&mut self, account_address: &str) -> Result<QueryScopesResponse> {
        let request = QueryScopesRequest {
            account_address: account_address.to_string(),
        };
        self.client
            .scopes(request)
            .await
            .map(|res| res.into_inner())
            .map_err(|e| Error::Query(e.to_string()))
    }

    pub async fn scopes_by_type(
        &mut self,
        account_address: &str,
        scope_type: i32,
    ) -> Result<QueryScopesByTypeResponse> {
        let request = QueryScopesByTypeRequest {
            account_address: account_address.to_string(),
            scope_type,
        };
        self.client
            .scopes_by_type(request)
            .await
            .map(|res| res.into_inner())
            .map_err(|e| Error::Query(e.to_string()))
    }

    pub async fn approved_clients(&mut self) -> Result<QueryApprovedClientsResponse> {
        let request = QueryApprovedClientsRequest {};
        self.client
            .approved_clients(request)
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
