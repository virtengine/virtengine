// @generated
/// Generated client implementations.
pub mod query_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    #[derive(Debug, Clone)]
    pub struct QueryClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl QueryClient<tonic::transport::Channel> {
        /// Attempt to create a new client by connecting to a given endpoint.
        pub async fn connect<D>(dst: D) -> Result<Self, tonic::transport::Error>
        where
            D: TryInto<tonic::transport::Endpoint>,
            D::Error: Into<StdError>,
        {
            let conn = tonic::transport::Endpoint::new(dst)?.connect().await?;
            Ok(Self::new(conn))
        }
    }
    impl<T> QueryClient<T>
    where
        T: tonic::client::GrpcService<tonic::body::BoxBody>,
        T::Error: Into<StdError>,
        T::ResponseBody: Body<Data = Bytes> + Send + 'static,
        <T::ResponseBody as Body>::Error: Into<StdError> + Send,
    {
        pub fn new(inner: T) -> Self {
            let inner = tonic::client::Grpc::new(inner);
            Self { inner }
        }
        pub fn with_origin(inner: T, origin: Uri) -> Self {
            let inner = tonic::client::Grpc::with_origin(inner, origin);
            Self { inner }
        }
        pub fn with_interceptor<F>(
            inner: T,
            interceptor: F,
        ) -> QueryClient<InterceptedService<T, F>>
        where
            F: tonic::service::Interceptor,
            T::ResponseBody: Default,
            T: tonic::codegen::Service<
                http::Request<tonic::body::BoxBody>,
                Response = http::Response<
                    <T as tonic::client::GrpcService<tonic::body::BoxBody>>::ResponseBody,
                >,
            >,
            <T as tonic::codegen::Service<
                http::Request<tonic::body::BoxBody>,
            >>::Error: Into<StdError> + Send + Sync,
        {
            QueryClient::new(InterceptedService::new(inner, interceptor))
        }
        /// Compress requests with the given encoding.
        ///
        /// This requires the server to support it otherwise it might respond with an
        /// error.
        #[must_use]
        pub fn send_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.inner = self.inner.send_compressed(encoding);
            self
        }
        /// Enable decompressing responses.
        #[must_use]
        pub fn accept_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.inner = self.inner.accept_compressed(encoding);
            self
        }
        /// Limits the maximum size of a decoded message.
        ///
        /// Default: `4MB`
        #[must_use]
        pub fn max_decoding_message_size(mut self, limit: usize) -> Self {
            self.inner = self.inner.max_decoding_message_size(limit);
            self
        }
        /// Limits the maximum size of an encoded message.
        ///
        /// Default: `usize::MAX`
        #[must_use]
        pub fn max_encoding_message_size(mut self, limit: usize) -> Self {
            self.inner = self.inner.max_encoding_message_size(limit);
            self
        }
        pub async fn identity_record(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryIdentityRecordRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryIdentityRecordResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/IdentityRecord",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "IdentityRecord"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn identity(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryIdentityRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryIdentityResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/Identity",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "Identity"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn scope(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryScopeRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryScopeResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/Scope",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "Scope"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn scopes(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryScopesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryScopesResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/Scopes",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "Scopes"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn scopes_by_type(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryScopesByTypeRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryScopesByTypeResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/ScopesByType",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "ScopesByType"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn identity_score(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryIdentityScoreRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryIdentityScoreResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/IdentityScore",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "IdentityScore"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn identity_status(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryIdentityStatusRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryIdentityStatusResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/IdentityStatus",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "IdentityStatus"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn identity_wallet(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryIdentityWalletRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryIdentityWalletResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/IdentityWallet",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "IdentityWallet"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn wallet_scopes(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryWalletScopesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryWalletScopesResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/WalletScopes",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "WalletScopes"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn consent_settings(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryConsentSettingsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryConsentSettingsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/ConsentSettings",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "ConsentSettings"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn derived_features(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryDerivedFeaturesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryDerivedFeaturesResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/DerivedFeatures",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "DerivedFeatures"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn derived_feature_hashes(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryDerivedFeatureHashesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryDerivedFeatureHashesResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/DerivedFeatureHashes",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.veid.v1.Query", "DerivedFeatureHashes"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn verification_history(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryVerificationHistoryRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryVerificationHistoryResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/VerificationHistory",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.veid.v1.Query", "VerificationHistory"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn approved_clients(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryApprovedClientsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryApprovedClientsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/ApprovedClients",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "ApprovedClients"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn params(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryParamsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryParamsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/Params",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "Params"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn borderline_params(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryBorderlineParamsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryBorderlineParamsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/BorderlineParams",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "BorderlineParams"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn appeal(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryAppealRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryAppealResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/Appeal",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "Appeal"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn appeals(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryAppealsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryAppealsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/Appeals",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "Appeals"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn appeals_by_scope(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryAppealsByScopeRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryAppealsByScopeResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/AppealsByScope",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "AppealsByScope"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn appeal_params(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryAppealParamsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryAppealParamsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/AppealParams",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "AppealParams"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn compliance_status(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryComplianceStatusRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryComplianceStatusResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/ComplianceStatus",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "ComplianceStatus"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn compliance_provider(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryComplianceProviderRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryComplianceProviderResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/ComplianceProvider",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.veid.v1.Query", "ComplianceProvider"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn compliance_providers(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryComplianceProvidersRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryComplianceProvidersResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/ComplianceProviders",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.veid.v1.Query", "ComplianceProviders"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn compliance_params(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryComplianceParamsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryComplianceParamsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/ComplianceParams",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "ComplianceParams"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn model_version(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryModelVersionRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryModelVersionResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/ModelVersion",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "ModelVersion"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn active_models(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryActiveModelsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryActiveModelsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/ActiveModels",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "ActiveModels"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn model_history(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryModelHistoryRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryModelHistoryResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/ModelHistory",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "ModelHistory"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn validator_model_sync(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryValidatorModelSyncRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryValidatorModelSyncResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/ValidatorModelSync",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.veid.v1.Query", "ValidatorModelSync"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn model_params(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryModelParamsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryModelParamsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Query/ModelParams",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Query", "ModelParams"));
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod query_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with QueryServer.
    #[async_trait]
    pub trait Query: Send + Sync + 'static {
        async fn identity_record(
            &self,
            request: tonic::Request<super::QueryIdentityRecordRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryIdentityRecordResponse>,
            tonic::Status,
        >;
        async fn identity(
            &self,
            request: tonic::Request<super::QueryIdentityRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryIdentityResponse>,
            tonic::Status,
        >;
        async fn scope(
            &self,
            request: tonic::Request<super::QueryScopeRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryScopeResponse>,
            tonic::Status,
        >;
        async fn scopes(
            &self,
            request: tonic::Request<super::QueryScopesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryScopesResponse>,
            tonic::Status,
        >;
        async fn scopes_by_type(
            &self,
            request: tonic::Request<super::QueryScopesByTypeRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryScopesByTypeResponse>,
            tonic::Status,
        >;
        async fn identity_score(
            &self,
            request: tonic::Request<super::QueryIdentityScoreRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryIdentityScoreResponse>,
            tonic::Status,
        >;
        async fn identity_status(
            &self,
            request: tonic::Request<super::QueryIdentityStatusRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryIdentityStatusResponse>,
            tonic::Status,
        >;
        async fn identity_wallet(
            &self,
            request: tonic::Request<super::QueryIdentityWalletRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryIdentityWalletResponse>,
            tonic::Status,
        >;
        async fn wallet_scopes(
            &self,
            request: tonic::Request<super::QueryWalletScopesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryWalletScopesResponse>,
            tonic::Status,
        >;
        async fn consent_settings(
            &self,
            request: tonic::Request<super::QueryConsentSettingsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryConsentSettingsResponse>,
            tonic::Status,
        >;
        async fn derived_features(
            &self,
            request: tonic::Request<super::QueryDerivedFeaturesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryDerivedFeaturesResponse>,
            tonic::Status,
        >;
        async fn derived_feature_hashes(
            &self,
            request: tonic::Request<super::QueryDerivedFeatureHashesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryDerivedFeatureHashesResponse>,
            tonic::Status,
        >;
        async fn verification_history(
            &self,
            request: tonic::Request<super::QueryVerificationHistoryRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryVerificationHistoryResponse>,
            tonic::Status,
        >;
        async fn approved_clients(
            &self,
            request: tonic::Request<super::QueryApprovedClientsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryApprovedClientsResponse>,
            tonic::Status,
        >;
        async fn params(
            &self,
            request: tonic::Request<super::QueryParamsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryParamsResponse>,
            tonic::Status,
        >;
        async fn borderline_params(
            &self,
            request: tonic::Request<super::QueryBorderlineParamsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryBorderlineParamsResponse>,
            tonic::Status,
        >;
        async fn appeal(
            &self,
            request: tonic::Request<super::QueryAppealRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryAppealResponse>,
            tonic::Status,
        >;
        async fn appeals(
            &self,
            request: tonic::Request<super::QueryAppealsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryAppealsResponse>,
            tonic::Status,
        >;
        async fn appeals_by_scope(
            &self,
            request: tonic::Request<super::QueryAppealsByScopeRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryAppealsByScopeResponse>,
            tonic::Status,
        >;
        async fn appeal_params(
            &self,
            request: tonic::Request<super::QueryAppealParamsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryAppealParamsResponse>,
            tonic::Status,
        >;
        async fn compliance_status(
            &self,
            request: tonic::Request<super::QueryComplianceStatusRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryComplianceStatusResponse>,
            tonic::Status,
        >;
        async fn compliance_provider(
            &self,
            request: tonic::Request<super::QueryComplianceProviderRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryComplianceProviderResponse>,
            tonic::Status,
        >;
        async fn compliance_providers(
            &self,
            request: tonic::Request<super::QueryComplianceProvidersRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryComplianceProvidersResponse>,
            tonic::Status,
        >;
        async fn compliance_params(
            &self,
            request: tonic::Request<super::QueryComplianceParamsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryComplianceParamsResponse>,
            tonic::Status,
        >;
        async fn model_version(
            &self,
            request: tonic::Request<super::QueryModelVersionRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryModelVersionResponse>,
            tonic::Status,
        >;
        async fn active_models(
            &self,
            request: tonic::Request<super::QueryActiveModelsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryActiveModelsResponse>,
            tonic::Status,
        >;
        async fn model_history(
            &self,
            request: tonic::Request<super::QueryModelHistoryRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryModelHistoryResponse>,
            tonic::Status,
        >;
        async fn validator_model_sync(
            &self,
            request: tonic::Request<super::QueryValidatorModelSyncRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryValidatorModelSyncResponse>,
            tonic::Status,
        >;
        async fn model_params(
            &self,
            request: tonic::Request<super::QueryModelParamsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryModelParamsResponse>,
            tonic::Status,
        >;
    }
    #[derive(Debug)]
    pub struct QueryServer<T: Query> {
        inner: Arc<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    impl<T: Query> QueryServer<T> {
        pub fn new(inner: T) -> Self {
            Self::from_arc(Arc::new(inner))
        }
        pub fn from_arc(inner: Arc<T>) -> Self {
            Self {
                inner,
                accept_compression_encodings: Default::default(),
                send_compression_encodings: Default::default(),
                max_decoding_message_size: None,
                max_encoding_message_size: None,
            }
        }
        pub fn with_interceptor<F>(
            inner: T,
            interceptor: F,
        ) -> InterceptedService<Self, F>
        where
            F: tonic::service::Interceptor,
        {
            InterceptedService::new(Self::new(inner), interceptor)
        }
        /// Enable decompressing requests with the given encoding.
        #[must_use]
        pub fn accept_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.accept_compression_encodings.enable(encoding);
            self
        }
        /// Compress responses with the given encoding, if the client supports it.
        #[must_use]
        pub fn send_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.send_compression_encodings.enable(encoding);
            self
        }
        /// Limits the maximum size of a decoded message.
        ///
        /// Default: `4MB`
        #[must_use]
        pub fn max_decoding_message_size(mut self, limit: usize) -> Self {
            self.max_decoding_message_size = Some(limit);
            self
        }
        /// Limits the maximum size of an encoded message.
        ///
        /// Default: `usize::MAX`
        #[must_use]
        pub fn max_encoding_message_size(mut self, limit: usize) -> Self {
            self.max_encoding_message_size = Some(limit);
            self
        }
    }
    impl<T, B> tonic::codegen::Service<http::Request<B>> for QueryServer<T>
    where
        T: Query,
        B: Body + Send + 'static,
        B::Error: Into<StdError> + Send + 'static,
    {
        type Response = http::Response<tonic::body::BoxBody>;
        type Error = std::convert::Infallible;
        type Future = BoxFuture<Self::Response, Self::Error>;
        fn poll_ready(
            &mut self,
            _cx: &mut Context<'_>,
        ) -> Poll<std::result::Result<(), Self::Error>> {
            Poll::Ready(Ok(()))
        }
        fn call(&mut self, req: http::Request<B>) -> Self::Future {
            match req.uri().path() {
                "/virtengine.veid.v1.Query/IdentityRecord" => {
                    #[allow(non_camel_case_types)]
                    struct IdentityRecordSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryIdentityRecordRequest>
                    for IdentityRecordSvc<T> {
                        type Response = super::QueryIdentityRecordResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryIdentityRecordRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::identity_record(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = IdentityRecordSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/Identity" => {
                    #[allow(non_camel_case_types)]
                    struct IdentitySvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryIdentityRequest>
                    for IdentitySvc<T> {
                        type Response = super::QueryIdentityResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryIdentityRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::identity(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = IdentitySvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/Scope" => {
                    #[allow(non_camel_case_types)]
                    struct ScopeSvc<T: Query>(pub Arc<T>);
                    impl<T: Query> tonic::server::UnaryService<super::QueryScopeRequest>
                    for ScopeSvc<T> {
                        type Response = super::QueryScopeResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryScopeRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::scope(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ScopeSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/Scopes" => {
                    #[allow(non_camel_case_types)]
                    struct ScopesSvc<T: Query>(pub Arc<T>);
                    impl<T: Query> tonic::server::UnaryService<super::QueryScopesRequest>
                    for ScopesSvc<T> {
                        type Response = super::QueryScopesResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryScopesRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::scopes(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ScopesSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/ScopesByType" => {
                    #[allow(non_camel_case_types)]
                    struct ScopesByTypeSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryScopesByTypeRequest>
                    for ScopesByTypeSvc<T> {
                        type Response = super::QueryScopesByTypeResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryScopesByTypeRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::scopes_by_type(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ScopesByTypeSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/IdentityScore" => {
                    #[allow(non_camel_case_types)]
                    struct IdentityScoreSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryIdentityScoreRequest>
                    for IdentityScoreSvc<T> {
                        type Response = super::QueryIdentityScoreResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryIdentityScoreRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::identity_score(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = IdentityScoreSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/IdentityStatus" => {
                    #[allow(non_camel_case_types)]
                    struct IdentityStatusSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryIdentityStatusRequest>
                    for IdentityStatusSvc<T> {
                        type Response = super::QueryIdentityStatusResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryIdentityStatusRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::identity_status(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = IdentityStatusSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/IdentityWallet" => {
                    #[allow(non_camel_case_types)]
                    struct IdentityWalletSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryIdentityWalletRequest>
                    for IdentityWalletSvc<T> {
                        type Response = super::QueryIdentityWalletResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryIdentityWalletRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::identity_wallet(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = IdentityWalletSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/WalletScopes" => {
                    #[allow(non_camel_case_types)]
                    struct WalletScopesSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryWalletScopesRequest>
                    for WalletScopesSvc<T> {
                        type Response = super::QueryWalletScopesResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryWalletScopesRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::wallet_scopes(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = WalletScopesSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/ConsentSettings" => {
                    #[allow(non_camel_case_types)]
                    struct ConsentSettingsSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryConsentSettingsRequest>
                    for ConsentSettingsSvc<T> {
                        type Response = super::QueryConsentSettingsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryConsentSettingsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::consent_settings(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ConsentSettingsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/DerivedFeatures" => {
                    #[allow(non_camel_case_types)]
                    struct DerivedFeaturesSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryDerivedFeaturesRequest>
                    for DerivedFeaturesSvc<T> {
                        type Response = super::QueryDerivedFeaturesResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryDerivedFeaturesRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::derived_features(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = DerivedFeaturesSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/DerivedFeatureHashes" => {
                    #[allow(non_camel_case_types)]
                    struct DerivedFeatureHashesSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<
                        super::QueryDerivedFeatureHashesRequest,
                    > for DerivedFeatureHashesSvc<T> {
                        type Response = super::QueryDerivedFeatureHashesResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::QueryDerivedFeatureHashesRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::derived_feature_hashes(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = DerivedFeatureHashesSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/VerificationHistory" => {
                    #[allow(non_camel_case_types)]
                    struct VerificationHistorySvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryVerificationHistoryRequest>
                    for VerificationHistorySvc<T> {
                        type Response = super::QueryVerificationHistoryResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::QueryVerificationHistoryRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::verification_history(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = VerificationHistorySvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/ApprovedClients" => {
                    #[allow(non_camel_case_types)]
                    struct ApprovedClientsSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryApprovedClientsRequest>
                    for ApprovedClientsSvc<T> {
                        type Response = super::QueryApprovedClientsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryApprovedClientsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::approved_clients(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ApprovedClientsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/Params" => {
                    #[allow(non_camel_case_types)]
                    struct ParamsSvc<T: Query>(pub Arc<T>);
                    impl<T: Query> tonic::server::UnaryService<super::QueryParamsRequest>
                    for ParamsSvc<T> {
                        type Response = super::QueryParamsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryParamsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::params(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ParamsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/BorderlineParams" => {
                    #[allow(non_camel_case_types)]
                    struct BorderlineParamsSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryBorderlineParamsRequest>
                    for BorderlineParamsSvc<T> {
                        type Response = super::QueryBorderlineParamsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryBorderlineParamsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::borderline_params(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = BorderlineParamsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/Appeal" => {
                    #[allow(non_camel_case_types)]
                    struct AppealSvc<T: Query>(pub Arc<T>);
                    impl<T: Query> tonic::server::UnaryService<super::QueryAppealRequest>
                    for AppealSvc<T> {
                        type Response = super::QueryAppealResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryAppealRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::appeal(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = AppealSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/Appeals" => {
                    #[allow(non_camel_case_types)]
                    struct AppealsSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryAppealsRequest>
                    for AppealsSvc<T> {
                        type Response = super::QueryAppealsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryAppealsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::appeals(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = AppealsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/AppealsByScope" => {
                    #[allow(non_camel_case_types)]
                    struct AppealsByScopeSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryAppealsByScopeRequest>
                    for AppealsByScopeSvc<T> {
                        type Response = super::QueryAppealsByScopeResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryAppealsByScopeRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::appeals_by_scope(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = AppealsByScopeSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/AppealParams" => {
                    #[allow(non_camel_case_types)]
                    struct AppealParamsSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryAppealParamsRequest>
                    for AppealParamsSvc<T> {
                        type Response = super::QueryAppealParamsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryAppealParamsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::appeal_params(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = AppealParamsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/ComplianceStatus" => {
                    #[allow(non_camel_case_types)]
                    struct ComplianceStatusSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryComplianceStatusRequest>
                    for ComplianceStatusSvc<T> {
                        type Response = super::QueryComplianceStatusResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryComplianceStatusRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::compliance_status(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ComplianceStatusSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/ComplianceProvider" => {
                    #[allow(non_camel_case_types)]
                    struct ComplianceProviderSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryComplianceProviderRequest>
                    for ComplianceProviderSvc<T> {
                        type Response = super::QueryComplianceProviderResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::QueryComplianceProviderRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::compliance_provider(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ComplianceProviderSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/ComplianceProviders" => {
                    #[allow(non_camel_case_types)]
                    struct ComplianceProvidersSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryComplianceProvidersRequest>
                    for ComplianceProvidersSvc<T> {
                        type Response = super::QueryComplianceProvidersResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::QueryComplianceProvidersRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::compliance_providers(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ComplianceProvidersSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/ComplianceParams" => {
                    #[allow(non_camel_case_types)]
                    struct ComplianceParamsSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryComplianceParamsRequest>
                    for ComplianceParamsSvc<T> {
                        type Response = super::QueryComplianceParamsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryComplianceParamsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::compliance_params(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ComplianceParamsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/ModelVersion" => {
                    #[allow(non_camel_case_types)]
                    struct ModelVersionSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryModelVersionRequest>
                    for ModelVersionSvc<T> {
                        type Response = super::QueryModelVersionResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryModelVersionRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::model_version(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ModelVersionSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/ActiveModels" => {
                    #[allow(non_camel_case_types)]
                    struct ActiveModelsSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryActiveModelsRequest>
                    for ActiveModelsSvc<T> {
                        type Response = super::QueryActiveModelsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryActiveModelsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::active_models(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ActiveModelsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/ModelHistory" => {
                    #[allow(non_camel_case_types)]
                    struct ModelHistorySvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryModelHistoryRequest>
                    for ModelHistorySvc<T> {
                        type Response = super::QueryModelHistoryResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryModelHistoryRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::model_history(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ModelHistorySvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/ValidatorModelSync" => {
                    #[allow(non_camel_case_types)]
                    struct ValidatorModelSyncSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryValidatorModelSyncRequest>
                    for ValidatorModelSyncSvc<T> {
                        type Response = super::QueryValidatorModelSyncResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::QueryValidatorModelSyncRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::validator_model_sync(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ValidatorModelSyncSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Query/ModelParams" => {
                    #[allow(non_camel_case_types)]
                    struct ModelParamsSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryModelParamsRequest>
                    for ModelParamsSvc<T> {
                        type Response = super::QueryModelParamsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryModelParamsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::model_params(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ModelParamsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                _ => {
                    Box::pin(async move {
                        Ok(
                            http::Response::builder()
                                .status(200)
                                .header("grpc-status", tonic::Code::Unimplemented as i32)
                                .header(
                                    http::header::CONTENT_TYPE,
                                    tonic::metadata::GRPC_CONTENT_TYPE,
                                )
                                .body(empty_body())
                                .unwrap(),
                        )
                    })
                }
            }
        }
    }
    impl<T: Query> Clone for QueryServer<T> {
        fn clone(&self) -> Self {
            let inner = self.inner.clone();
            Self {
                inner,
                accept_compression_encodings: self.accept_compression_encodings,
                send_compression_encodings: self.send_compression_encodings,
                max_decoding_message_size: self.max_decoding_message_size,
                max_encoding_message_size: self.max_encoding_message_size,
            }
        }
    }
    impl<T: Query> tonic::server::NamedService for QueryServer<T> {
        const NAME: &'static str = "virtengine.veid.v1.Query";
    }
}
/// Generated client implementations.
pub mod msg_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    #[derive(Debug, Clone)]
    pub struct MsgClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl MsgClient<tonic::transport::Channel> {
        /// Attempt to create a new client by connecting to a given endpoint.
        pub async fn connect<D>(dst: D) -> Result<Self, tonic::transport::Error>
        where
            D: TryInto<tonic::transport::Endpoint>,
            D::Error: Into<StdError>,
        {
            let conn = tonic::transport::Endpoint::new(dst)?.connect().await?;
            Ok(Self::new(conn))
        }
    }
    impl<T> MsgClient<T>
    where
        T: tonic::client::GrpcService<tonic::body::BoxBody>,
        T::Error: Into<StdError>,
        T::ResponseBody: Body<Data = Bytes> + Send + 'static,
        <T::ResponseBody as Body>::Error: Into<StdError> + Send,
    {
        pub fn new(inner: T) -> Self {
            let inner = tonic::client::Grpc::new(inner);
            Self { inner }
        }
        pub fn with_origin(inner: T, origin: Uri) -> Self {
            let inner = tonic::client::Grpc::with_origin(inner, origin);
            Self { inner }
        }
        pub fn with_interceptor<F>(
            inner: T,
            interceptor: F,
        ) -> MsgClient<InterceptedService<T, F>>
        where
            F: tonic::service::Interceptor,
            T::ResponseBody: Default,
            T: tonic::codegen::Service<
                http::Request<tonic::body::BoxBody>,
                Response = http::Response<
                    <T as tonic::client::GrpcService<tonic::body::BoxBody>>::ResponseBody,
                >,
            >,
            <T as tonic::codegen::Service<
                http::Request<tonic::body::BoxBody>,
            >>::Error: Into<StdError> + Send + Sync,
        {
            MsgClient::new(InterceptedService::new(inner, interceptor))
        }
        /// Compress requests with the given encoding.
        ///
        /// This requires the server to support it otherwise it might respond with an
        /// error.
        #[must_use]
        pub fn send_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.inner = self.inner.send_compressed(encoding);
            self
        }
        /// Enable decompressing responses.
        #[must_use]
        pub fn accept_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.inner = self.inner.accept_compressed(encoding);
            self
        }
        /// Limits the maximum size of a decoded message.
        ///
        /// Default: `4MB`
        #[must_use]
        pub fn max_decoding_message_size(mut self, limit: usize) -> Self {
            self.inner = self.inner.max_decoding_message_size(limit);
            self
        }
        /// Limits the maximum size of an encoded message.
        ///
        /// Default: `usize::MAX`
        #[must_use]
        pub fn max_encoding_message_size(mut self, limit: usize) -> Self {
            self.inner = self.inner.max_encoding_message_size(limit);
            self
        }
        pub async fn upload_scope(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgUploadScope>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUploadScopeResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/UploadScope",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "UploadScope"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn revoke_scope(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgRevokeScope>,
        ) -> std::result::Result<
            tonic::Response<super::MsgRevokeScopeResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/RevokeScope",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "RevokeScope"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn request_verification(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgRequestVerification>,
        ) -> std::result::Result<
            tonic::Response<super::MsgRequestVerificationResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/RequestVerification",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.veid.v1.Msg", "RequestVerification"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn update_verification_status(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgUpdateVerificationStatus>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateVerificationStatusResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/UpdateVerificationStatus",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.veid.v1.Msg", "UpdateVerificationStatus"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn update_score(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgUpdateScore>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateScoreResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/UpdateScore",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "UpdateScore"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn create_identity_wallet(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgCreateIdentityWallet>,
        ) -> std::result::Result<
            tonic::Response<super::MsgCreateIdentityWalletResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/CreateIdentityWallet",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.veid.v1.Msg", "CreateIdentityWallet"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn add_scope_to_wallet(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgAddScopeToWallet>,
        ) -> std::result::Result<
            tonic::Response<super::MsgAddScopeToWalletResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/AddScopeToWallet",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "AddScopeToWallet"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn revoke_scope_from_wallet(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgRevokeScopeFromWallet>,
        ) -> std::result::Result<
            tonic::Response<super::MsgRevokeScopeFromWalletResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/RevokeScopeFromWallet",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.veid.v1.Msg", "RevokeScopeFromWallet"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn update_consent_settings(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgUpdateConsentSettings>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateConsentSettingsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/UpdateConsentSettings",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.veid.v1.Msg", "UpdateConsentSettings"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn rebind_wallet(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgRebindWallet>,
        ) -> std::result::Result<
            tonic::Response<super::MsgRebindWalletResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/RebindWallet",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "RebindWallet"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn update_derived_features(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgUpdateDerivedFeatures>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateDerivedFeaturesResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/UpdateDerivedFeatures",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.veid.v1.Msg", "UpdateDerivedFeatures"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn complete_borderline_fallback(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgCompleteBorderlineFallback>,
        ) -> std::result::Result<
            tonic::Response<super::MsgCompleteBorderlineFallbackResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/CompleteBorderlineFallback",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "virtengine.veid.v1.Msg",
                        "CompleteBorderlineFallback",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn update_borderline_params(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgUpdateBorderlineParams>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateBorderlineParamsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/UpdateBorderlineParams",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.veid.v1.Msg", "UpdateBorderlineParams"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn update_params(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgUpdateParams>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateParamsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/UpdateParams",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "UpdateParams"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn submit_appeal(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgSubmitAppeal>,
        ) -> std::result::Result<
            tonic::Response<super::MsgSubmitAppealResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/SubmitAppeal",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "SubmitAppeal"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn claim_appeal(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgClaimAppeal>,
        ) -> std::result::Result<
            tonic::Response<super::MsgClaimAppealResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/ClaimAppeal",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "ClaimAppeal"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn resolve_appeal(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgResolveAppeal>,
        ) -> std::result::Result<
            tonic::Response<super::MsgResolveAppealResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/ResolveAppeal",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "ResolveAppeal"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn withdraw_appeal(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgWithdrawAppeal>,
        ) -> std::result::Result<
            tonic::Response<super::MsgWithdrawAppealResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/WithdrawAppeal",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "WithdrawAppeal"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn submit_compliance_check(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgSubmitComplianceCheck>,
        ) -> std::result::Result<
            tonic::Response<super::MsgSubmitComplianceCheckResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/SubmitComplianceCheck",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.veid.v1.Msg", "SubmitComplianceCheck"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn attest_compliance(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgAttestCompliance>,
        ) -> std::result::Result<
            tonic::Response<super::MsgAttestComplianceResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/AttestCompliance",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "AttestCompliance"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn update_compliance_params(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgUpdateComplianceParams>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateComplianceParamsResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/UpdateComplianceParams",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.veid.v1.Msg", "UpdateComplianceParams"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn register_compliance_provider(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgRegisterComplianceProvider>,
        ) -> std::result::Result<
            tonic::Response<super::MsgRegisterComplianceProviderResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/RegisterComplianceProvider",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "virtengine.veid.v1.Msg",
                        "RegisterComplianceProvider",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn deactivate_compliance_provider(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgDeactivateComplianceProvider>,
        ) -> std::result::Result<
            tonic::Response<super::MsgDeactivateComplianceProviderResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/DeactivateComplianceProvider",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "virtengine.veid.v1.Msg",
                        "DeactivateComplianceProvider",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn register_model(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgRegisterModel>,
        ) -> std::result::Result<
            tonic::Response<super::MsgRegisterModelResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/RegisterModel",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "RegisterModel"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn propose_model_update(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgProposeModelUpdate>,
        ) -> std::result::Result<
            tonic::Response<super::MsgProposeModelUpdateResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/ProposeModelUpdate",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "ProposeModelUpdate"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn report_model_version(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgReportModelVersion>,
        ) -> std::result::Result<
            tonic::Response<super::MsgReportModelVersionResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/ReportModelVersion",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "ReportModelVersion"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn activate_model(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgActivateModel>,
        ) -> std::result::Result<
            tonic::Response<super::MsgActivateModelResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/ActivateModel",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "ActivateModel"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn deprecate_model(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgDeprecateModel>,
        ) -> std::result::Result<
            tonic::Response<super::MsgDeprecateModelResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/DeprecateModel",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "DeprecateModel"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn revoke_model(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgRevokeModel>,
        ) -> std::result::Result<
            tonic::Response<super::MsgRevokeModelResponse>,
            tonic::Status,
        > {
            self.inner
                .ready()
                .await
                .map_err(|e| {
                    tonic::Status::new(
                        tonic::Code::Unknown,
                        format!("Service was not ready: {}", e.into()),
                    )
                })?;
            let codec = tonic::codec::ProstCodec::default();
            let path = http::uri::PathAndQuery::from_static(
                "/virtengine.veid.v1.Msg/RevokeModel",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.veid.v1.Msg", "RevokeModel"));
            self.inner.unary(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod msg_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with MsgServer.
    #[async_trait]
    pub trait Msg: Send + Sync + 'static {
        async fn upload_scope(
            &self,
            request: tonic::Request<super::MsgUploadScope>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUploadScopeResponse>,
            tonic::Status,
        >;
        async fn revoke_scope(
            &self,
            request: tonic::Request<super::MsgRevokeScope>,
        ) -> std::result::Result<
            tonic::Response<super::MsgRevokeScopeResponse>,
            tonic::Status,
        >;
        async fn request_verification(
            &self,
            request: tonic::Request<super::MsgRequestVerification>,
        ) -> std::result::Result<
            tonic::Response<super::MsgRequestVerificationResponse>,
            tonic::Status,
        >;
        async fn update_verification_status(
            &self,
            request: tonic::Request<super::MsgUpdateVerificationStatus>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateVerificationStatusResponse>,
            tonic::Status,
        >;
        async fn update_score(
            &self,
            request: tonic::Request<super::MsgUpdateScore>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateScoreResponse>,
            tonic::Status,
        >;
        async fn create_identity_wallet(
            &self,
            request: tonic::Request<super::MsgCreateIdentityWallet>,
        ) -> std::result::Result<
            tonic::Response<super::MsgCreateIdentityWalletResponse>,
            tonic::Status,
        >;
        async fn add_scope_to_wallet(
            &self,
            request: tonic::Request<super::MsgAddScopeToWallet>,
        ) -> std::result::Result<
            tonic::Response<super::MsgAddScopeToWalletResponse>,
            tonic::Status,
        >;
        async fn revoke_scope_from_wallet(
            &self,
            request: tonic::Request<super::MsgRevokeScopeFromWallet>,
        ) -> std::result::Result<
            tonic::Response<super::MsgRevokeScopeFromWalletResponse>,
            tonic::Status,
        >;
        async fn update_consent_settings(
            &self,
            request: tonic::Request<super::MsgUpdateConsentSettings>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateConsentSettingsResponse>,
            tonic::Status,
        >;
        async fn rebind_wallet(
            &self,
            request: tonic::Request<super::MsgRebindWallet>,
        ) -> std::result::Result<
            tonic::Response<super::MsgRebindWalletResponse>,
            tonic::Status,
        >;
        async fn update_derived_features(
            &self,
            request: tonic::Request<super::MsgUpdateDerivedFeatures>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateDerivedFeaturesResponse>,
            tonic::Status,
        >;
        async fn complete_borderline_fallback(
            &self,
            request: tonic::Request<super::MsgCompleteBorderlineFallback>,
        ) -> std::result::Result<
            tonic::Response<super::MsgCompleteBorderlineFallbackResponse>,
            tonic::Status,
        >;
        async fn update_borderline_params(
            &self,
            request: tonic::Request<super::MsgUpdateBorderlineParams>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateBorderlineParamsResponse>,
            tonic::Status,
        >;
        async fn update_params(
            &self,
            request: tonic::Request<super::MsgUpdateParams>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateParamsResponse>,
            tonic::Status,
        >;
        async fn submit_appeal(
            &self,
            request: tonic::Request<super::MsgSubmitAppeal>,
        ) -> std::result::Result<
            tonic::Response<super::MsgSubmitAppealResponse>,
            tonic::Status,
        >;
        async fn claim_appeal(
            &self,
            request: tonic::Request<super::MsgClaimAppeal>,
        ) -> std::result::Result<
            tonic::Response<super::MsgClaimAppealResponse>,
            tonic::Status,
        >;
        async fn resolve_appeal(
            &self,
            request: tonic::Request<super::MsgResolveAppeal>,
        ) -> std::result::Result<
            tonic::Response<super::MsgResolveAppealResponse>,
            tonic::Status,
        >;
        async fn withdraw_appeal(
            &self,
            request: tonic::Request<super::MsgWithdrawAppeal>,
        ) -> std::result::Result<
            tonic::Response<super::MsgWithdrawAppealResponse>,
            tonic::Status,
        >;
        async fn submit_compliance_check(
            &self,
            request: tonic::Request<super::MsgSubmitComplianceCheck>,
        ) -> std::result::Result<
            tonic::Response<super::MsgSubmitComplianceCheckResponse>,
            tonic::Status,
        >;
        async fn attest_compliance(
            &self,
            request: tonic::Request<super::MsgAttestCompliance>,
        ) -> std::result::Result<
            tonic::Response<super::MsgAttestComplianceResponse>,
            tonic::Status,
        >;
        async fn update_compliance_params(
            &self,
            request: tonic::Request<super::MsgUpdateComplianceParams>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateComplianceParamsResponse>,
            tonic::Status,
        >;
        async fn register_compliance_provider(
            &self,
            request: tonic::Request<super::MsgRegisterComplianceProvider>,
        ) -> std::result::Result<
            tonic::Response<super::MsgRegisterComplianceProviderResponse>,
            tonic::Status,
        >;
        async fn deactivate_compliance_provider(
            &self,
            request: tonic::Request<super::MsgDeactivateComplianceProvider>,
        ) -> std::result::Result<
            tonic::Response<super::MsgDeactivateComplianceProviderResponse>,
            tonic::Status,
        >;
        async fn register_model(
            &self,
            request: tonic::Request<super::MsgRegisterModel>,
        ) -> std::result::Result<
            tonic::Response<super::MsgRegisterModelResponse>,
            tonic::Status,
        >;
        async fn propose_model_update(
            &self,
            request: tonic::Request<super::MsgProposeModelUpdate>,
        ) -> std::result::Result<
            tonic::Response<super::MsgProposeModelUpdateResponse>,
            tonic::Status,
        >;
        async fn report_model_version(
            &self,
            request: tonic::Request<super::MsgReportModelVersion>,
        ) -> std::result::Result<
            tonic::Response<super::MsgReportModelVersionResponse>,
            tonic::Status,
        >;
        async fn activate_model(
            &self,
            request: tonic::Request<super::MsgActivateModel>,
        ) -> std::result::Result<
            tonic::Response<super::MsgActivateModelResponse>,
            tonic::Status,
        >;
        async fn deprecate_model(
            &self,
            request: tonic::Request<super::MsgDeprecateModel>,
        ) -> std::result::Result<
            tonic::Response<super::MsgDeprecateModelResponse>,
            tonic::Status,
        >;
        async fn revoke_model(
            &self,
            request: tonic::Request<super::MsgRevokeModel>,
        ) -> std::result::Result<
            tonic::Response<super::MsgRevokeModelResponse>,
            tonic::Status,
        >;
    }
    #[derive(Debug)]
    pub struct MsgServer<T: Msg> {
        inner: Arc<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    impl<T: Msg> MsgServer<T> {
        pub fn new(inner: T) -> Self {
            Self::from_arc(Arc::new(inner))
        }
        pub fn from_arc(inner: Arc<T>) -> Self {
            Self {
                inner,
                accept_compression_encodings: Default::default(),
                send_compression_encodings: Default::default(),
                max_decoding_message_size: None,
                max_encoding_message_size: None,
            }
        }
        pub fn with_interceptor<F>(
            inner: T,
            interceptor: F,
        ) -> InterceptedService<Self, F>
        where
            F: tonic::service::Interceptor,
        {
            InterceptedService::new(Self::new(inner), interceptor)
        }
        /// Enable decompressing requests with the given encoding.
        #[must_use]
        pub fn accept_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.accept_compression_encodings.enable(encoding);
            self
        }
        /// Compress responses with the given encoding, if the client supports it.
        #[must_use]
        pub fn send_compressed(mut self, encoding: CompressionEncoding) -> Self {
            self.send_compression_encodings.enable(encoding);
            self
        }
        /// Limits the maximum size of a decoded message.
        ///
        /// Default: `4MB`
        #[must_use]
        pub fn max_decoding_message_size(mut self, limit: usize) -> Self {
            self.max_decoding_message_size = Some(limit);
            self
        }
        /// Limits the maximum size of an encoded message.
        ///
        /// Default: `usize::MAX`
        #[must_use]
        pub fn max_encoding_message_size(mut self, limit: usize) -> Self {
            self.max_encoding_message_size = Some(limit);
            self
        }
    }
    impl<T, B> tonic::codegen::Service<http::Request<B>> for MsgServer<T>
    where
        T: Msg,
        B: Body + Send + 'static,
        B::Error: Into<StdError> + Send + 'static,
    {
        type Response = http::Response<tonic::body::BoxBody>;
        type Error = std::convert::Infallible;
        type Future = BoxFuture<Self::Response, Self::Error>;
        fn poll_ready(
            &mut self,
            _cx: &mut Context<'_>,
        ) -> Poll<std::result::Result<(), Self::Error>> {
            Poll::Ready(Ok(()))
        }
        fn call(&mut self, req: http::Request<B>) -> Self::Future {
            match req.uri().path() {
                "/virtengine.veid.v1.Msg/UploadScope" => {
                    #[allow(non_camel_case_types)]
                    struct UploadScopeSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgUploadScope>
                    for UploadScopeSvc<T> {
                        type Response = super::MsgUploadScopeResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgUploadScope>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::upload_scope(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = UploadScopeSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/RevokeScope" => {
                    #[allow(non_camel_case_types)]
                    struct RevokeScopeSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgRevokeScope>
                    for RevokeScopeSvc<T> {
                        type Response = super::MsgRevokeScopeResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgRevokeScope>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::revoke_scope(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = RevokeScopeSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/RequestVerification" => {
                    #[allow(non_camel_case_types)]
                    struct RequestVerificationSvc<T: Msg>(pub Arc<T>);
                    impl<
                        T: Msg,
                    > tonic::server::UnaryService<super::MsgRequestVerification>
                    for RequestVerificationSvc<T> {
                        type Response = super::MsgRequestVerificationResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgRequestVerification>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::request_verification(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = RequestVerificationSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/UpdateVerificationStatus" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateVerificationStatusSvc<T: Msg>(pub Arc<T>);
                    impl<
                        T: Msg,
                    > tonic::server::UnaryService<super::MsgUpdateVerificationStatus>
                    for UpdateVerificationStatusSvc<T> {
                        type Response = super::MsgUpdateVerificationStatusResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgUpdateVerificationStatus>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::update_verification_status(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = UpdateVerificationStatusSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/UpdateScore" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateScoreSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgUpdateScore>
                    for UpdateScoreSvc<T> {
                        type Response = super::MsgUpdateScoreResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgUpdateScore>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::update_score(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = UpdateScoreSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/CreateIdentityWallet" => {
                    #[allow(non_camel_case_types)]
                    struct CreateIdentityWalletSvc<T: Msg>(pub Arc<T>);
                    impl<
                        T: Msg,
                    > tonic::server::UnaryService<super::MsgCreateIdentityWallet>
                    for CreateIdentityWalletSvc<T> {
                        type Response = super::MsgCreateIdentityWalletResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgCreateIdentityWallet>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::create_identity_wallet(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = CreateIdentityWalletSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/AddScopeToWallet" => {
                    #[allow(non_camel_case_types)]
                    struct AddScopeToWalletSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgAddScopeToWallet>
                    for AddScopeToWalletSvc<T> {
                        type Response = super::MsgAddScopeToWalletResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgAddScopeToWallet>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::add_scope_to_wallet(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = AddScopeToWalletSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/RevokeScopeFromWallet" => {
                    #[allow(non_camel_case_types)]
                    struct RevokeScopeFromWalletSvc<T: Msg>(pub Arc<T>);
                    impl<
                        T: Msg,
                    > tonic::server::UnaryService<super::MsgRevokeScopeFromWallet>
                    for RevokeScopeFromWalletSvc<T> {
                        type Response = super::MsgRevokeScopeFromWalletResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgRevokeScopeFromWallet>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::revoke_scope_from_wallet(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = RevokeScopeFromWalletSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/UpdateConsentSettings" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateConsentSettingsSvc<T: Msg>(pub Arc<T>);
                    impl<
                        T: Msg,
                    > tonic::server::UnaryService<super::MsgUpdateConsentSettings>
                    for UpdateConsentSettingsSvc<T> {
                        type Response = super::MsgUpdateConsentSettingsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgUpdateConsentSettings>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::update_consent_settings(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = UpdateConsentSettingsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/RebindWallet" => {
                    #[allow(non_camel_case_types)]
                    struct RebindWalletSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgRebindWallet>
                    for RebindWalletSvc<T> {
                        type Response = super::MsgRebindWalletResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgRebindWallet>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::rebind_wallet(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = RebindWalletSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/UpdateDerivedFeatures" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateDerivedFeaturesSvc<T: Msg>(pub Arc<T>);
                    impl<
                        T: Msg,
                    > tonic::server::UnaryService<super::MsgUpdateDerivedFeatures>
                    for UpdateDerivedFeaturesSvc<T> {
                        type Response = super::MsgUpdateDerivedFeaturesResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgUpdateDerivedFeatures>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::update_derived_features(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = UpdateDerivedFeaturesSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/CompleteBorderlineFallback" => {
                    #[allow(non_camel_case_types)]
                    struct CompleteBorderlineFallbackSvc<T: Msg>(pub Arc<T>);
                    impl<
                        T: Msg,
                    > tonic::server::UnaryService<super::MsgCompleteBorderlineFallback>
                    for CompleteBorderlineFallbackSvc<T> {
                        type Response = super::MsgCompleteBorderlineFallbackResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgCompleteBorderlineFallback>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::complete_borderline_fallback(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = CompleteBorderlineFallbackSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/UpdateBorderlineParams" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateBorderlineParamsSvc<T: Msg>(pub Arc<T>);
                    impl<
                        T: Msg,
                    > tonic::server::UnaryService<super::MsgUpdateBorderlineParams>
                    for UpdateBorderlineParamsSvc<T> {
                        type Response = super::MsgUpdateBorderlineParamsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgUpdateBorderlineParams>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::update_borderline_params(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = UpdateBorderlineParamsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/UpdateParams" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateParamsSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgUpdateParams>
                    for UpdateParamsSvc<T> {
                        type Response = super::MsgUpdateParamsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgUpdateParams>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::update_params(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = UpdateParamsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/SubmitAppeal" => {
                    #[allow(non_camel_case_types)]
                    struct SubmitAppealSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgSubmitAppeal>
                    for SubmitAppealSvc<T> {
                        type Response = super::MsgSubmitAppealResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgSubmitAppeal>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::submit_appeal(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = SubmitAppealSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/ClaimAppeal" => {
                    #[allow(non_camel_case_types)]
                    struct ClaimAppealSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgClaimAppeal>
                    for ClaimAppealSvc<T> {
                        type Response = super::MsgClaimAppealResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgClaimAppeal>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::claim_appeal(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ClaimAppealSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/ResolveAppeal" => {
                    #[allow(non_camel_case_types)]
                    struct ResolveAppealSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgResolveAppeal>
                    for ResolveAppealSvc<T> {
                        type Response = super::MsgResolveAppealResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgResolveAppeal>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::resolve_appeal(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ResolveAppealSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/WithdrawAppeal" => {
                    #[allow(non_camel_case_types)]
                    struct WithdrawAppealSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgWithdrawAppeal>
                    for WithdrawAppealSvc<T> {
                        type Response = super::MsgWithdrawAppealResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgWithdrawAppeal>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::withdraw_appeal(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = WithdrawAppealSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/SubmitComplianceCheck" => {
                    #[allow(non_camel_case_types)]
                    struct SubmitComplianceCheckSvc<T: Msg>(pub Arc<T>);
                    impl<
                        T: Msg,
                    > tonic::server::UnaryService<super::MsgSubmitComplianceCheck>
                    for SubmitComplianceCheckSvc<T> {
                        type Response = super::MsgSubmitComplianceCheckResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgSubmitComplianceCheck>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::submit_compliance_check(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = SubmitComplianceCheckSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/AttestCompliance" => {
                    #[allow(non_camel_case_types)]
                    struct AttestComplianceSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgAttestCompliance>
                    for AttestComplianceSvc<T> {
                        type Response = super::MsgAttestComplianceResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgAttestCompliance>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::attest_compliance(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = AttestComplianceSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/UpdateComplianceParams" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateComplianceParamsSvc<T: Msg>(pub Arc<T>);
                    impl<
                        T: Msg,
                    > tonic::server::UnaryService<super::MsgUpdateComplianceParams>
                    for UpdateComplianceParamsSvc<T> {
                        type Response = super::MsgUpdateComplianceParamsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgUpdateComplianceParams>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::update_compliance_params(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = UpdateComplianceParamsSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/RegisterComplianceProvider" => {
                    #[allow(non_camel_case_types)]
                    struct RegisterComplianceProviderSvc<T: Msg>(pub Arc<T>);
                    impl<
                        T: Msg,
                    > tonic::server::UnaryService<super::MsgRegisterComplianceProvider>
                    for RegisterComplianceProviderSvc<T> {
                        type Response = super::MsgRegisterComplianceProviderResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgRegisterComplianceProvider>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::register_compliance_provider(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = RegisterComplianceProviderSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/DeactivateComplianceProvider" => {
                    #[allow(non_camel_case_types)]
                    struct DeactivateComplianceProviderSvc<T: Msg>(pub Arc<T>);
                    impl<
                        T: Msg,
                    > tonic::server::UnaryService<super::MsgDeactivateComplianceProvider>
                    for DeactivateComplianceProviderSvc<T> {
                        type Response = super::MsgDeactivateComplianceProviderResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::MsgDeactivateComplianceProvider,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::deactivate_compliance_provider(&inner, request)
                                    .await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = DeactivateComplianceProviderSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/RegisterModel" => {
                    #[allow(non_camel_case_types)]
                    struct RegisterModelSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgRegisterModel>
                    for RegisterModelSvc<T> {
                        type Response = super::MsgRegisterModelResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgRegisterModel>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::register_model(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = RegisterModelSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/ProposeModelUpdate" => {
                    #[allow(non_camel_case_types)]
                    struct ProposeModelUpdateSvc<T: Msg>(pub Arc<T>);
                    impl<
                        T: Msg,
                    > tonic::server::UnaryService<super::MsgProposeModelUpdate>
                    for ProposeModelUpdateSvc<T> {
                        type Response = super::MsgProposeModelUpdateResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgProposeModelUpdate>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::propose_model_update(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ProposeModelUpdateSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/ReportModelVersion" => {
                    #[allow(non_camel_case_types)]
                    struct ReportModelVersionSvc<T: Msg>(pub Arc<T>);
                    impl<
                        T: Msg,
                    > tonic::server::UnaryService<super::MsgReportModelVersion>
                    for ReportModelVersionSvc<T> {
                        type Response = super::MsgReportModelVersionResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgReportModelVersion>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::report_model_version(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ReportModelVersionSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/ActivateModel" => {
                    #[allow(non_camel_case_types)]
                    struct ActivateModelSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgActivateModel>
                    for ActivateModelSvc<T> {
                        type Response = super::MsgActivateModelResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgActivateModel>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::activate_model(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = ActivateModelSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/DeprecateModel" => {
                    #[allow(non_camel_case_types)]
                    struct DeprecateModelSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgDeprecateModel>
                    for DeprecateModelSvc<T> {
                        type Response = super::MsgDeprecateModelResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgDeprecateModel>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::deprecate_model(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = DeprecateModelSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.veid.v1.Msg/RevokeModel" => {
                    #[allow(non_camel_case_types)]
                    struct RevokeModelSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgRevokeModel>
                    for RevokeModelSvc<T> {
                        type Response = super::MsgRevokeModelResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgRevokeModel>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::revoke_model(&inner, request).await
                            };
                            Box::pin(fut)
                        }
                    }
                    let accept_compression_encodings = self.accept_compression_encodings;
                    let send_compression_encodings = self.send_compression_encodings;
                    let max_decoding_message_size = self.max_decoding_message_size;
                    let max_encoding_message_size = self.max_encoding_message_size;
                    let inner = self.inner.clone();
                    let fut = async move {
                        let method = RevokeModelSvc(inner);
                        let codec = tonic::codec::ProstCodec::default();
                        let mut grpc = tonic::server::Grpc::new(codec)
                            .apply_compression_config(
                                accept_compression_encodings,
                                send_compression_encodings,
                            )
                            .apply_max_message_size_config(
                                max_decoding_message_size,
                                max_encoding_message_size,
                            );
                        let res = grpc.unary(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                _ => {
                    Box::pin(async move {
                        Ok(
                            http::Response::builder()
                                .status(200)
                                .header("grpc-status", tonic::Code::Unimplemented as i32)
                                .header(
                                    http::header::CONTENT_TYPE,
                                    tonic::metadata::GRPC_CONTENT_TYPE,
                                )
                                .body(empty_body())
                                .unwrap(),
                        )
                    })
                }
            }
        }
    }
    impl<T: Msg> Clone for MsgServer<T> {
        fn clone(&self) -> Self {
            let inner = self.inner.clone();
            Self {
                inner,
                accept_compression_encodings: self.accept_compression_encodings,
                send_compression_encodings: self.send_compression_encodings,
                max_decoding_message_size: self.max_decoding_message_size,
                max_encoding_message_size: self.max_encoding_message_size,
            }
        }
    }
    impl<T: Msg> tonic::server::NamedService for MsgServer<T> {
        const NAME: &'static str = "virtengine.veid.v1.Msg";
    }
}
