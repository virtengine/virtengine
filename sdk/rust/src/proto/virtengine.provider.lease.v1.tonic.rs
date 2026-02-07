// @generated
/// Generated client implementations.
pub mod lease_rpc_client {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    use tonic::codegen::http::Uri;
    /** LeaseRPC defines the RPC server for lease control
*/
    #[derive(Debug, Clone)]
    pub struct LeaseRpcClient<T> {
        inner: tonic::client::Grpc<T>,
    }
    impl LeaseRpcClient<tonic::transport::Channel> {
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
    impl<T> LeaseRpcClient<T>
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
        ) -> LeaseRpcClient<InterceptedService<T, F>>
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
            LeaseRpcClient::new(InterceptedService::new(inner, interceptor))
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
        /** SendManifest sends manifest to the provider
*/
        pub async fn send_manifest(
            &mut self,
            request: impl tonic::IntoRequest<super::SendManifestRequest>,
        ) -> std::result::Result<
            tonic::Response<super::SendManifestResponse>,
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
                "/virtengine.provider.lease.v1.LeaseRPC/SendManifest",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "virtengine.provider.lease.v1.LeaseRPC",
                        "SendManifest",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** ServiceStatus
 buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE
 buf:lint:ignore RPC_RESPONSE_STANDARD_NAME
*/
        pub async fn service_status(
            &mut self,
            request: impl tonic::IntoRequest<super::ServiceStatusRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ServiceStatusResponse>,
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
                "/virtengine.provider.lease.v1.LeaseRPC/ServiceStatus",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "virtengine.provider.lease.v1.LeaseRPC",
                        "ServiceStatus",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** StreamServiceStatus
 buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE
 buf:lint:ignore RPC_RESPONSE_STANDARD_NAME
*/
        pub async fn stream_service_status(
            &mut self,
            request: impl tonic::IntoRequest<super::ServiceStatusRequest>,
        ) -> std::result::Result<
            tonic::Response<tonic::codec::Streaming<super::ServiceStatusResponse>>,
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
                "/virtengine.provider.lease.v1.LeaseRPC/StreamServiceStatus",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "virtengine.provider.lease.v1.LeaseRPC",
                        "StreamServiceStatus",
                    ),
                );
            self.inner.server_streaming(req, path, codec).await
        }
        /** ServiceLogs
 buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE
 buf:lint:ignore RPC_RESPONSE_STANDARD_NAME
*/
        pub async fn service_logs(
            &mut self,
            request: impl tonic::IntoRequest<super::ServiceLogsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ServiceLogsResponse>,
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
                "/virtengine.provider.lease.v1.LeaseRPC/ServiceLogs",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "virtengine.provider.lease.v1.LeaseRPC",
                        "ServiceLogs",
                    ),
                );
            self.inner.unary(req, path, codec).await
        }
        /** StreamServiceLogs
 buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE
 buf:lint:ignore RPC_RESPONSE_STANDARD_NAME
*/
        pub async fn stream_service_logs(
            &mut self,
            request: impl tonic::IntoRequest<super::ServiceLogsRequest>,
        ) -> std::result::Result<
            tonic::Response<tonic::codec::Streaming<super::ServiceLogsResponse>>,
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
                "/virtengine.provider.lease.v1.LeaseRPC/StreamServiceLogs",
            );
            let mut req = request.into_request();
            req.extensions_mut()
                .insert(
                    GrpcMethod::new(
                        "virtengine.provider.lease.v1.LeaseRPC",
                        "StreamServiceLogs",
                    ),
                );
            self.inner.server_streaming(req, path, codec).await
        }
    }
}
/// Generated server implementations.
pub mod lease_rpc_server {
    #![allow(unused_variables, dead_code, missing_docs, clippy::let_unit_value)]
    use tonic::codegen::*;
    /// Generated trait containing gRPC methods that should be implemented for use with LeaseRpcServer.
    #[async_trait]
    pub trait LeaseRpc: Send + Sync + 'static {
        /** SendManifest sends manifest to the provider
*/
        async fn send_manifest(
            &self,
            request: tonic::Request<super::SendManifestRequest>,
        ) -> std::result::Result<
            tonic::Response<super::SendManifestResponse>,
            tonic::Status,
        >;
        /** ServiceStatus
 buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE
 buf:lint:ignore RPC_RESPONSE_STANDARD_NAME
*/
        async fn service_status(
            &self,
            request: tonic::Request<super::ServiceStatusRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ServiceStatusResponse>,
            tonic::Status,
        >;
        /// Server streaming response type for the StreamServiceStatus method.
        type StreamServiceStatusStream: tonic::codegen::tokio_stream::Stream<
                Item = std::result::Result<super::ServiceStatusResponse, tonic::Status>,
            >
            + Send
            + 'static;
        /** StreamServiceStatus
 buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE
 buf:lint:ignore RPC_RESPONSE_STANDARD_NAME
*/
        async fn stream_service_status(
            &self,
            request: tonic::Request<super::ServiceStatusRequest>,
        ) -> std::result::Result<
            tonic::Response<Self::StreamServiceStatusStream>,
            tonic::Status,
        >;
        /** ServiceLogs
 buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE
 buf:lint:ignore RPC_RESPONSE_STANDARD_NAME
*/
        async fn service_logs(
            &self,
            request: tonic::Request<super::ServiceLogsRequest>,
        ) -> std::result::Result<
            tonic::Response<super::ServiceLogsResponse>,
            tonic::Status,
        >;
        /// Server streaming response type for the StreamServiceLogs method.
        type StreamServiceLogsStream: tonic::codegen::tokio_stream::Stream<
                Item = std::result::Result<super::ServiceLogsResponse, tonic::Status>,
            >
            + Send
            + 'static;
        /** StreamServiceLogs
 buf:lint:ignore RPC_REQUEST_RESPONSE_UNIQUE
 buf:lint:ignore RPC_RESPONSE_STANDARD_NAME
*/
        async fn stream_service_logs(
            &self,
            request: tonic::Request<super::ServiceLogsRequest>,
        ) -> std::result::Result<
            tonic::Response<Self::StreamServiceLogsStream>,
            tonic::Status,
        >;
    }
    /** LeaseRPC defines the RPC server for lease control
*/
    #[derive(Debug)]
    pub struct LeaseRpcServer<T: LeaseRpc> {
        inner: Arc<T>,
        accept_compression_encodings: EnabledCompressionEncodings,
        send_compression_encodings: EnabledCompressionEncodings,
        max_decoding_message_size: Option<usize>,
        max_encoding_message_size: Option<usize>,
    }
    impl<T: LeaseRpc> LeaseRpcServer<T> {
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
    impl<T, B> tonic::codegen::Service<http::Request<B>> for LeaseRpcServer<T>
    where
        T: LeaseRpc,
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
                "/virtengine.provider.lease.v1.LeaseRPC/SendManifest" => {
                    #[allow(non_camel_case_types)]
                    struct SendManifestSvc<T: LeaseRpc>(pub Arc<T>);
                    impl<
                        T: LeaseRpc,
                    > tonic::server::UnaryService<super::SendManifestRequest>
                    for SendManifestSvc<T> {
                        type Response = super::SendManifestResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::SendManifestRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as LeaseRpc>::send_manifest(&inner, request).await
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
                        let method = SendManifestSvc(inner);
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
                "/virtengine.provider.lease.v1.LeaseRPC/ServiceStatus" => {
                    #[allow(non_camel_case_types)]
                    struct ServiceStatusSvc<T: LeaseRpc>(pub Arc<T>);
                    impl<
                        T: LeaseRpc,
                    > tonic::server::UnaryService<super::ServiceStatusRequest>
                    for ServiceStatusSvc<T> {
                        type Response = super::ServiceStatusResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ServiceStatusRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as LeaseRpc>::service_status(&inner, request).await
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
                        let method = ServiceStatusSvc(inner);
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
                "/virtengine.provider.lease.v1.LeaseRPC/StreamServiceStatus" => {
                    #[allow(non_camel_case_types)]
                    struct StreamServiceStatusSvc<T: LeaseRpc>(pub Arc<T>);
                    impl<
                        T: LeaseRpc,
                    > tonic::server::ServerStreamingService<super::ServiceStatusRequest>
                    for StreamServiceStatusSvc<T> {
                        type Response = super::ServiceStatusResponse;
                        type ResponseStream = T::StreamServiceStatusStream;
                        type Future = BoxFuture<
                            tonic::Response<Self::ResponseStream>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ServiceStatusRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as LeaseRpc>::stream_service_status(&inner, request)
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
                        let method = StreamServiceStatusSvc(inner);
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
                        let res = grpc.server_streaming(method, req).await;
                        Ok(res)
                    };
                    Box::pin(fut)
                }
                "/virtengine.provider.lease.v1.LeaseRPC/ServiceLogs" => {
                    #[allow(non_camel_case_types)]
                    struct ServiceLogsSvc<T: LeaseRpc>(pub Arc<T>);
                    impl<
                        T: LeaseRpc,
                    > tonic::server::UnaryService<super::ServiceLogsRequest>
                    for ServiceLogsSvc<T> {
                        type Response = super::ServiceLogsResponse;
                        type Future = BoxFuture<
                            tonic::Response<Self::Response>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ServiceLogsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as LeaseRpc>::service_logs(&inner, request).await
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
                        let method = ServiceLogsSvc(inner);
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
                "/virtengine.provider.lease.v1.LeaseRPC/StreamServiceLogs" => {
                    #[allow(non_camel_case_types)]
                    struct StreamServiceLogsSvc<T: LeaseRpc>(pub Arc<T>);
                    impl<
                        T: LeaseRpc,
                    > tonic::server::ServerStreamingService<super::ServiceLogsRequest>
                    for StreamServiceLogsSvc<T> {
                        type Response = super::ServiceLogsResponse;
                        type ResponseStream = T::StreamServiceLogsStream;
                        type Future = BoxFuture<
                            tonic::Response<Self::ResponseStream>,
                            tonic::Status,
                        >;
                        fn call(
                            &mut self,
                            request: tonic::Request<super::ServiceLogsRequest>,
                        ) -> Self::Future {
                            let inner = Arc::clone(&self.0);
                            let fut = async move {
                                <T as LeaseRpc>::stream_service_logs(&inner, request).await
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
                        let method = StreamServiceLogsSvc(inner);
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
                        let res = grpc.server_streaming(method, req).await;
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
    impl<T: LeaseRpc> Clone for LeaseRpcServer<T> {
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
    impl<T: LeaseRpc> tonic::server::NamedService for LeaseRpcServer<T> {
        const NAME: &'static str = "virtengine.provider.lease.v1.LeaseRPC";
    }
}
