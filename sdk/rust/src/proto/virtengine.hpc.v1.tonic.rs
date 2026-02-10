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
        pub async fn cluster(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryClusterRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryClusterResponse>,
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
                "/virtengine.hpc.v1.Query/Cluster",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Query", "Cluster"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn clusters(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryClustersRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryClustersResponse>,
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
                "/virtengine.hpc.v1.Query/Clusters",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Query", "Clusters"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn clusters_by_provider(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryClustersByProviderRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryClustersByProviderResponse>,
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
                "/virtengine.hpc.v1.Query/ClustersByProvider",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.hpc.v1.Query", "ClustersByProvider"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn offering(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryOfferingRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryOfferingResponse>,
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
                "/virtengine.hpc.v1.Query/Offering",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Query", "Offering"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn offerings(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryOfferingsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryOfferingsResponse>,
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
                "/virtengine.hpc.v1.Query/Offerings",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Query", "Offerings"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn offerings_by_cluster(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryOfferingsByClusterRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryOfferingsByClusterResponse>,
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
                "/virtengine.hpc.v1.Query/OfferingsByCluster",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.hpc.v1.Query", "OfferingsByCluster"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn job(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryJobRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryJobResponse>,
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
                "/virtengine.hpc.v1.Query/Job",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Query", "Job"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn jobs(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryJobsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryJobsResponse>,
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
                "/virtengine.hpc.v1.Query/Jobs",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Query", "Jobs"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn jobs_by_customer(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryJobsByCustomerRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryJobsByCustomerResponse>,
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
                "/virtengine.hpc.v1.Query/JobsByCustomer",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Query", "JobsByCustomer"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn jobs_by_provider(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryJobsByProviderRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryJobsByProviderResponse>,
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
                "/virtengine.hpc.v1.Query/JobsByProvider",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Query", "JobsByProvider"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn job_accounting(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryJobAccountingRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryJobAccountingResponse>,
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
                "/virtengine.hpc.v1.Query/JobAccounting",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Query", "JobAccounting"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn node_metadata(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryNodeMetadataRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryNodeMetadataResponse>,
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
                "/virtengine.hpc.v1.Query/NodeMetadata",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Query", "NodeMetadata"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn nodes_by_cluster(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryNodesByClusterRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryNodesByClusterResponse>,
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
                "/virtengine.hpc.v1.Query/NodesByCluster",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Query", "NodesByCluster"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn scheduling_decision(
            &mut self,
            request: impl tonic::IntoRequest<super::QuerySchedulingDecisionRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QuerySchedulingDecisionResponse>,
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
                "/virtengine.hpc.v1.Query/SchedulingDecision",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.hpc.v1.Query", "SchedulingDecision"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn scheduling_decision_by_job(
            &mut self,
            request: impl tonic::IntoRequest<super::QuerySchedulingDecisionByJobRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QuerySchedulingDecisionByJobResponse>,
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
                "/virtengine.hpc.v1.Query/SchedulingDecisionByJob",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new("virtengine.hpc.v1.Query", "SchedulingDecisionByJob"),
                );
            self.inner.unary(req, path, codec).await
        }
        pub async fn reward(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryRewardRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryRewardResponse>,
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
                "/virtengine.hpc.v1.Query/Reward",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Query", "Reward"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn rewards_by_job(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryRewardsByJobRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryRewardsByJobResponse>,
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
                "/virtengine.hpc.v1.Query/RewardsByJob",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Query", "RewardsByJob"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn dispute(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryDisputeRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryDisputeResponse>,
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
                "/virtengine.hpc.v1.Query/Dispute",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Query", "Dispute"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn disputes(
            &mut self,
            request: impl tonic::IntoRequest<super::QueryDisputesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryDisputesResponse>,
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
                "/virtengine.hpc.v1.Query/Disputes",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Query", "Disputes"));
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
                "/virtengine.hpc.v1.Query/Params",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Query", "Params"));
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
        async fn cluster(
            &self,
            request: tonic::Request<super::QueryClusterRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryClusterResponse>,
            tonic::Status,
        >;
        async fn clusters(
            &self,
            request: tonic::Request<super::QueryClustersRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryClustersResponse>,
            tonic::Status,
        >;
        async fn clusters_by_provider(
            &self,
            request: tonic::Request<super::QueryClustersByProviderRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryClustersByProviderResponse>,
            tonic::Status,
        >;
        async fn offering(
            &self,
            request: tonic::Request<super::QueryOfferingRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryOfferingResponse>,
            tonic::Status,
        >;
        async fn offerings(
            &self,
            request: tonic::Request<super::QueryOfferingsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryOfferingsResponse>,
            tonic::Status,
        >;
        async fn offerings_by_cluster(
            &self,
            request: tonic::Request<super::QueryOfferingsByClusterRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryOfferingsByClusterResponse>,
            tonic::Status,
        >;
        async fn job(
            &self,
            request: tonic::Request<super::QueryJobRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryJobResponse>,
            tonic::Status,
        >;
        async fn jobs(
            &self,
            request: tonic::Request<super::QueryJobsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryJobsResponse>,
            tonic::Status,
        >;
        async fn jobs_by_customer(
            &self,
            request: tonic::Request<super::QueryJobsByCustomerRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryJobsByCustomerResponse>,
            tonic::Status,
        >;
        async fn jobs_by_provider(
            &self,
            request: tonic::Request<super::QueryJobsByProviderRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryJobsByProviderResponse>,
            tonic::Status,
        >;
        async fn job_accounting(
            &self,
            request: tonic::Request<super::QueryJobAccountingRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryJobAccountingResponse>,
            tonic::Status,
        >;
        async fn node_metadata(
            &self,
            request: tonic::Request<super::QueryNodeMetadataRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryNodeMetadataResponse>,
            tonic::Status,
        >;
        async fn nodes_by_cluster(
            &self,
            request: tonic::Request<super::QueryNodesByClusterRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryNodesByClusterResponse>,
            tonic::Status,
        >;
        async fn scheduling_decision(
            &self,
            request: tonic::Request<super::QuerySchedulingDecisionRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QuerySchedulingDecisionResponse>,
            tonic::Status,
        >;
        async fn scheduling_decision_by_job(
            &self,
            request: tonic::Request<super::QuerySchedulingDecisionByJobRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QuerySchedulingDecisionByJobResponse>,
            tonic::Status,
        >;
        async fn reward(
            &self,
            request: tonic::Request<super::QueryRewardRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryRewardResponse>,
            tonic::Status,
        >;
        async fn rewards_by_job(
            &self,
            request: tonic::Request<super::QueryRewardsByJobRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryRewardsByJobResponse>,
            tonic::Status,
        >;
        async fn dispute(
            &self,
            request: tonic::Request<super::QueryDisputeRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryDisputeResponse>,
            tonic::Status,
        >;
        async fn disputes(
            &self,
            request: tonic::Request<super::QueryDisputesRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryDisputesResponse>,
            tonic::Status,
        >;
        async fn params(
            &self,
            request: tonic::Request<super::QueryParamsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::QueryParamsResponse>,
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
                "/virtengine.hpc.v1.Query/Cluster" => {
                    #[allow(non_camel_case_types)]
                    struct ClusterSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryClusterRequest>
                    for ClusterSvc<T> {
                        type Response = super::QueryClusterResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryClusterRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::cluster(&inner, request).await
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
                        let method = ClusterSvc(inner);
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
                "/virtengine.hpc.v1.Query/Clusters" => {
                    #[allow(non_camel_case_types)]
                    struct ClustersSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryClustersRequest>
                    for ClustersSvc<T> {
                        type Response = super::QueryClustersResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryClustersRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::clusters(&inner, request).await
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
                        let method = ClustersSvc(inner);
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
                "/virtengine.hpc.v1.Query/ClustersByProvider" => {
                    #[allow(non_camel_case_types)]
                    struct ClustersByProviderSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryClustersByProviderRequest>
                    for ClustersByProviderSvc<T> {
                        type Response = super::QueryClustersByProviderResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::QueryClustersByProviderRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::clusters_by_provider(&inner, request).await
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
                        let method = ClustersByProviderSvc(inner);
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
                "/virtengine.hpc.v1.Query/Offering" => {
                    #[allow(non_camel_case_types)]
                    struct OfferingSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryOfferingRequest>
                    for OfferingSvc<T> {
                        type Response = super::QueryOfferingResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryOfferingRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::offering(&inner, request).await
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
                        let method = OfferingSvc(inner);
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
                "/virtengine.hpc.v1.Query/Offerings" => {
                    #[allow(non_camel_case_types)]
                    struct OfferingsSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryOfferingsRequest>
                    for OfferingsSvc<T> {
                        type Response = super::QueryOfferingsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryOfferingsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::offerings(&inner, request).await
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
                        let method = OfferingsSvc(inner);
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
                "/virtengine.hpc.v1.Query/OfferingsByCluster" => {
                    #[allow(non_camel_case_types)]
                    struct OfferingsByClusterSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryOfferingsByClusterRequest>
                    for OfferingsByClusterSvc<T> {
                        type Response = super::QueryOfferingsByClusterResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::QueryOfferingsByClusterRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::offerings_by_cluster(&inner, request).await
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
                        let method = OfferingsByClusterSvc(inner);
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
                "/virtengine.hpc.v1.Query/Job" => {
                    #[allow(non_camel_case_types)]
                    struct JobSvc<T: Query>(pub Arc<T>);
                    impl<T: Query> tonic::server::UnaryService<super::QueryJobRequest>
                    for JobSvc<T> {
                        type Response = super::QueryJobResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryJobRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::job(&inner, request).await
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
                        let method = JobSvc(inner);
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
                "/virtengine.hpc.v1.Query/Jobs" => {
                    #[allow(non_camel_case_types)]
                    struct JobsSvc<T: Query>(pub Arc<T>);
                    impl<T: Query> tonic::server::UnaryService<super::QueryJobsRequest>
                    for JobsSvc<T> {
                        type Response = super::QueryJobsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryJobsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::jobs(&inner, request).await
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
                        let method = JobsSvc(inner);
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
                "/virtengine.hpc.v1.Query/JobsByCustomer" => {
                    #[allow(non_camel_case_types)]
                    struct JobsByCustomerSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryJobsByCustomerRequest>
                    for JobsByCustomerSvc<T> {
                        type Response = super::QueryJobsByCustomerResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryJobsByCustomerRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::jobs_by_customer(&inner, request).await
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
                        let method = JobsByCustomerSvc(inner);
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
                "/virtengine.hpc.v1.Query/JobsByProvider" => {
                    #[allow(non_camel_case_types)]
                    struct JobsByProviderSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryJobsByProviderRequest>
                    for JobsByProviderSvc<T> {
                        type Response = super::QueryJobsByProviderResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryJobsByProviderRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::jobs_by_provider(&inner, request).await
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
                        let method = JobsByProviderSvc(inner);
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
                "/virtengine.hpc.v1.Query/JobAccounting" => {
                    #[allow(non_camel_case_types)]
                    struct JobAccountingSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryJobAccountingRequest>
                    for JobAccountingSvc<T> {
                        type Response = super::QueryJobAccountingResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryJobAccountingRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::job_accounting(&inner, request).await
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
                        let method = JobAccountingSvc(inner);
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
                "/virtengine.hpc.v1.Query/NodeMetadata" => {
                    #[allow(non_camel_case_types)]
                    struct NodeMetadataSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryNodeMetadataRequest>
                    for NodeMetadataSvc<T> {
                        type Response = super::QueryNodeMetadataResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryNodeMetadataRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::node_metadata(&inner, request).await
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
                        let method = NodeMetadataSvc(inner);
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
                "/virtengine.hpc.v1.Query/NodesByCluster" => {
                    #[allow(non_camel_case_types)]
                    struct NodesByClusterSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryNodesByClusterRequest>
                    for NodesByClusterSvc<T> {
                        type Response = super::QueryNodesByClusterResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryNodesByClusterRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::nodes_by_cluster(&inner, request).await
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
                        let method = NodesByClusterSvc(inner);
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
                "/virtengine.hpc.v1.Query/SchedulingDecision" => {
                    #[allow(non_camel_case_types)]
                    struct SchedulingDecisionSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QuerySchedulingDecisionRequest>
                    for SchedulingDecisionSvc<T> {
                        type Response = super::QuerySchedulingDecisionResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::QuerySchedulingDecisionRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::scheduling_decision(&inner, request).await
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
                        let method = SchedulingDecisionSvc(inner);
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
                "/virtengine.hpc.v1.Query/SchedulingDecisionByJob" => {
                    #[allow(non_camel_case_types)]
                    struct SchedulingDecisionByJobSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<
                        super::QuerySchedulingDecisionByJobRequest,
                    > for SchedulingDecisionByJobSvc<T> {
                        type Response = super::QuerySchedulingDecisionByJobResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<
                                super::QuerySchedulingDecisionByJobRequest,
                            >,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::scheduling_decision_by_job(&inner, request)
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
                        let method = SchedulingDecisionByJobSvc(inner);
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
                "/virtengine.hpc.v1.Query/Reward" => {
                    #[allow(non_camel_case_types)]
                    struct RewardSvc<T: Query>(pub Arc<T>);
                    impl<T: Query> tonic::server::UnaryService<super::QueryRewardRequest>
                    for RewardSvc<T> {
                        type Response = super::QueryRewardResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryRewardRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::reward(&inner, request).await
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
                        let method = RewardSvc(inner);
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
                "/virtengine.hpc.v1.Query/RewardsByJob" => {
                    #[allow(non_camel_case_types)]
                    struct RewardsByJobSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryRewardsByJobRequest>
                    for RewardsByJobSvc<T> {
                        type Response = super::QueryRewardsByJobResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryRewardsByJobRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::rewards_by_job(&inner, request).await
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
                        let method = RewardsByJobSvc(inner);
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
                "/virtengine.hpc.v1.Query/Dispute" => {
                    #[allow(non_camel_case_types)]
                    struct DisputeSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryDisputeRequest>
                    for DisputeSvc<T> {
                        type Response = super::QueryDisputeResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryDisputeRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::dispute(&inner, request).await
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
                        let method = DisputeSvc(inner);
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
                "/virtengine.hpc.v1.Query/Disputes" => {
                    #[allow(non_camel_case_types)]
                    struct DisputesSvc<T: Query>(pub Arc<T>);
                    impl<
                        T: Query,
                    > tonic::server::UnaryService<super::QueryDisputesRequest>
                    for DisputesSvc<T> {
                        type Response = super::QueryDisputesResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::QueryDisputesRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Query>::disputes(&inner, request).await
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
                        let method = DisputesSvc(inner);
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
                "/virtengine.hpc.v1.Query/Params" => {
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
        const NAME: &'static str = "virtengine.hpc.v1.Query";
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
        pub async fn register_cluster(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgRegisterCluster>,
        ) -> std::result::Result<
            tonic::Response<super::MsgRegisterClusterResponse>,
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
                "/virtengine.hpc.v1.Msg/RegisterCluster",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Msg", "RegisterCluster"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn update_cluster(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgUpdateCluster>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateClusterResponse>,
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
                "/virtengine.hpc.v1.Msg/UpdateCluster",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Msg", "UpdateCluster"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn deregister_cluster(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgDeregisterCluster>,
        ) -> std::result::Result<
            tonic::Response<super::MsgDeregisterClusterResponse>,
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
                "/virtengine.hpc.v1.Msg/DeregisterCluster",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Msg", "DeregisterCluster"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn create_offering(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgCreateOffering>,
        ) -> std::result::Result<
            tonic::Response<super::MsgCreateOfferingResponse>,
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
                "/virtengine.hpc.v1.Msg/CreateOffering",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Msg", "CreateOffering"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn update_offering(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgUpdateOffering>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateOfferingResponse>,
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
                "/virtengine.hpc.v1.Msg/UpdateOffering",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Msg", "UpdateOffering"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn submit_job(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgSubmitJob>,
        ) -> std::result::Result<
            tonic::Response<super::MsgSubmitJobResponse>,
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
                "/virtengine.hpc.v1.Msg/SubmitJob",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Msg", "SubmitJob"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn cancel_job(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgCancelJob>,
        ) -> std::result::Result<
            tonic::Response<super::MsgCancelJobResponse>,
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
                "/virtengine.hpc.v1.Msg/CancelJob",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Msg", "CancelJob"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn report_job_status(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgReportJobStatus>,
        ) -> std::result::Result<
            tonic::Response<super::MsgReportJobStatusResponse>,
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
                "/virtengine.hpc.v1.Msg/ReportJobStatus",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Msg", "ReportJobStatus"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn update_node_metadata(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgUpdateNodeMetadata>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateNodeMetadataResponse>,
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
                "/virtengine.hpc.v1.Msg/UpdateNodeMetadata",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Msg", "UpdateNodeMetadata"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn flag_dispute(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgFlagDispute>,
        ) -> std::result::Result<
            tonic::Response<super::MsgFlagDisputeResponse>,
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
                "/virtengine.hpc.v1.Msg/FlagDispute",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Msg", "FlagDispute"));
            self.inner.unary(req, path, codec).await
        }
        pub async fn resolve_dispute(
            &mut self,
            request: impl tonic::IntoRequest<super::MsgResolveDispute>,
        ) -> std::result::Result<
            tonic::Response<super::MsgResolveDisputeResponse>,
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
                "/virtengine.hpc.v1.Msg/ResolveDispute",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Msg", "ResolveDispute"));
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
                "/virtengine.hpc.v1.Msg/UpdateParams",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(GrpcMethod::new("virtengine.hpc.v1.Msg", "UpdateParams"));
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
        async fn register_cluster(
            &self,
            request: tonic::Request<super::MsgRegisterCluster>,
        ) -> std::result::Result<
            tonic::Response<super::MsgRegisterClusterResponse>,
            tonic::Status,
        >;
        async fn update_cluster(
            &self,
            request: tonic::Request<super::MsgUpdateCluster>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateClusterResponse>,
            tonic::Status,
        >;
        async fn deregister_cluster(
            &self,
            request: tonic::Request<super::MsgDeregisterCluster>,
        ) -> std::result::Result<
            tonic::Response<super::MsgDeregisterClusterResponse>,
            tonic::Status,
        >;
        async fn create_offering(
            &self,
            request: tonic::Request<super::MsgCreateOffering>,
        ) -> std::result::Result<
            tonic::Response<super::MsgCreateOfferingResponse>,
            tonic::Status,
        >;
        async fn update_offering(
            &self,
            request: tonic::Request<super::MsgUpdateOffering>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateOfferingResponse>,
            tonic::Status,
        >;
        async fn submit_job(
            &self,
            request: tonic::Request<super::MsgSubmitJob>,
        ) -> std::result::Result<
            tonic::Response<super::MsgSubmitJobResponse>,
            tonic::Status,
        >;
        async fn cancel_job(
            &self,
            request: tonic::Request<super::MsgCancelJob>,
        ) -> std::result::Result<
            tonic::Response<super::MsgCancelJobResponse>,
            tonic::Status,
        >;
        async fn report_job_status(
            &self,
            request: tonic::Request<super::MsgReportJobStatus>,
        ) -> std::result::Result<
            tonic::Response<super::MsgReportJobStatusResponse>,
            tonic::Status,
        >;
        async fn update_node_metadata(
            &self,
            request: tonic::Request<super::MsgUpdateNodeMetadata>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateNodeMetadataResponse>,
            tonic::Status,
        >;
        async fn flag_dispute(
            &self,
            request: tonic::Request<super::MsgFlagDispute>,
        ) -> std::result::Result<
            tonic::Response<super::MsgFlagDisputeResponse>,
            tonic::Status,
        >;
        async fn resolve_dispute(
            &self,
            request: tonic::Request<super::MsgResolveDispute>,
        ) -> std::result::Result<
            tonic::Response<super::MsgResolveDisputeResponse>,
            tonic::Status,
        >;
        async fn update_params(
            &self,
            request: tonic::Request<super::MsgUpdateParams>,
        ) -> std::result::Result<
            tonic::Response<super::MsgUpdateParamsResponse>,
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
                "/virtengine.hpc.v1.Msg/RegisterCluster" => {
                    #[allow(non_camel_case_types)]
                    struct RegisterClusterSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgRegisterCluster>
                    for RegisterClusterSvc<T> {
                        type Response = super::MsgRegisterClusterResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgRegisterCluster>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::register_cluster(&inner, request).await
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
                        let method = RegisterClusterSvc(inner);
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
                "/virtengine.hpc.v1.Msg/UpdateCluster" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateClusterSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgUpdateCluster>
                    for UpdateClusterSvc<T> {
                        type Response = super::MsgUpdateClusterResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgUpdateCluster>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::update_cluster(&inner, request).await
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
                        let method = UpdateClusterSvc(inner);
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
                "/virtengine.hpc.v1.Msg/DeregisterCluster" => {
                    #[allow(non_camel_case_types)]
                    struct DeregisterClusterSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgDeregisterCluster>
                    for DeregisterClusterSvc<T> {
                        type Response = super::MsgDeregisterClusterResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgDeregisterCluster>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::deregister_cluster(&inner, request).await
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
                        let method = DeregisterClusterSvc(inner);
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
                "/virtengine.hpc.v1.Msg/CreateOffering" => {
                    #[allow(non_camel_case_types)]
                    struct CreateOfferingSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgCreateOffering>
                    for CreateOfferingSvc<T> {
                        type Response = super::MsgCreateOfferingResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgCreateOffering>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::create_offering(&inner, request).await
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
                        let method = CreateOfferingSvc(inner);
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
                "/virtengine.hpc.v1.Msg/UpdateOffering" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateOfferingSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgUpdateOffering>
                    for UpdateOfferingSvc<T> {
                        type Response = super::MsgUpdateOfferingResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgUpdateOffering>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::update_offering(&inner, request).await
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
                        let method = UpdateOfferingSvc(inner);
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
                "/virtengine.hpc.v1.Msg/SubmitJob" => {
                    #[allow(non_camel_case_types)]
                    struct SubmitJobSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgSubmitJob>
                    for SubmitJobSvc<T> {
                        type Response = super::MsgSubmitJobResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgSubmitJob>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::submit_job(&inner, request).await
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
                        let method = SubmitJobSvc(inner);
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
                "/virtengine.hpc.v1.Msg/CancelJob" => {
                    #[allow(non_camel_case_types)]
                    struct CancelJobSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgCancelJob>
                    for CancelJobSvc<T> {
                        type Response = super::MsgCancelJobResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgCancelJob>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::cancel_job(&inner, request).await
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
                        let method = CancelJobSvc(inner);
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
                "/virtengine.hpc.v1.Msg/ReportJobStatus" => {
                    #[allow(non_camel_case_types)]
                    struct ReportJobStatusSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgReportJobStatus>
                    for ReportJobStatusSvc<T> {
                        type Response = super::MsgReportJobStatusResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgReportJobStatus>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::report_job_status(&inner, request).await
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
                        let method = ReportJobStatusSvc(inner);
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
                "/virtengine.hpc.v1.Msg/UpdateNodeMetadata" => {
                    #[allow(non_camel_case_types)]
                    struct UpdateNodeMetadataSvc<T: Msg>(pub Arc<T>);
                    impl<
                        T: Msg,
                    > tonic::server::UnaryService<super::MsgUpdateNodeMetadata>
                    for UpdateNodeMetadataSvc<T> {
                        type Response = super::MsgUpdateNodeMetadataResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgUpdateNodeMetadata>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::update_node_metadata(&inner, request).await
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
                        let method = UpdateNodeMetadataSvc(inner);
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
                "/virtengine.hpc.v1.Msg/FlagDispute" => {
                    #[allow(non_camel_case_types)]
                    struct FlagDisputeSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgFlagDispute>
                    for FlagDisputeSvc<T> {
                        type Response = super::MsgFlagDisputeResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgFlagDispute>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::flag_dispute(&inner, request).await
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
                        let method = FlagDisputeSvc(inner);
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
                "/virtengine.hpc.v1.Msg/ResolveDispute" => {
                    #[allow(non_camel_case_types)]
                    struct ResolveDisputeSvc<T: Msg>(pub Arc<T>);
                    impl<T: Msg> tonic::server::UnaryService<super::MsgResolveDispute>
                    for ResolveDisputeSvc<T> {
                        type Response = super::MsgResolveDisputeResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::MsgResolveDispute>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as Msg>::resolve_dispute(&inner, request).await
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
                        let method = ResolveDisputeSvc(inner);
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
                "/virtengine.hpc.v1.Msg/UpdateParams" => {
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
        const NAME: &'static str = "virtengine.hpc.v1.Msg";
    }
}
