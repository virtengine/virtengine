import { createServiceLoader } from "../sdk/client/createServiceLoader.ts";
import { SDKOptions } from "../sdk/types.ts";

import type * as google_protobuf_empty from "./protos/google/protobuf/empty.ts";
import type * as virtengine_provider_lease_v1_service from "./protos/virtengine/provider/lease/v1/service.ts";
import { createClientFactory } from "../sdk/client/createClientFactory.ts";
import type { Transport, CallOptions } from "../sdk/transport/types.ts";
import { withMetadata } from "../sdk/client/sdkMetadata.ts";
import type { DeepPartial } from "../encoding/typeEncodingHelpers.ts";


export const serviceLoader= createServiceLoader([
  () => import("./protos/virtengine/inventory/v1/service_virtengine.ts").then(m => m.NodeRPC),
  () => import("./protos/virtengine/inventory/v1/service_virtengine.ts").then(m => m.ClusterRPC),
  () => import("./protos/virtengine/provider/lease/v1/service_virtengine.ts").then(m => m.LeaseRPC),
  () => import("./protos/virtengine/provider/v1/service_virtengine.ts").then(m => m.ProviderRPC)
] as const);
export function createSDK(transport: Transport, options?: SDKOptions) {
  const getClient = createClientFactory<CallOptions>(transport, options?.clientOptions);
  return {
    virtengine: {
      inventory: {
        v1: {
          /**
           * queryNode defines a method to query hardware state of the node
           */
          queryNode: withMetadata(async function queryNode(input: DeepPartial<google_protobuf_empty.Empty> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(0);
            return getClient(service).queryNode(input, options);
          }, { path: [0, 0] }),
          /**
           * streamNode defines a method to stream hardware state of the node
           */
          streamNode: withMetadata(async function streamNode(input: DeepPartial<google_protobuf_empty.Empty> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(0);
            return getClient(service).streamNode(input, options);
          }, { path: [0, 1] }),
          /**
           * queryCluster defines a method to query hardware state of the cluster
           */
          queryCluster: withMetadata(async function queryCluster(input: DeepPartial<google_protobuf_empty.Empty> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getClient(service).queryCluster(input, options);
          }, { path: [1, 0] }),
          /**
           * streamCluster defines a method to stream hardware state of the cluster
           */
          streamCluster: withMetadata(async function streamCluster(input: DeepPartial<google_protobuf_empty.Empty> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(1);
            return getClient(service).streamCluster(input, options);
          }, { path: [1, 1] })
        }
      },
      provider: {
        lease: {
          v1: {
            /**
             * sendManifest sends manifest to the provider
             */
            sendManifest: withMetadata(async function sendManifest(input: DeepPartial<virtengine_provider_lease_v1_service.SendManifestRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(2);
              return getClient(service).sendManifest(input, options);
            }, { path: [2, 0] }),
            /**
             * serviceStatus
             */
            serviceStatus: withMetadata(async function serviceStatus(input: DeepPartial<virtengine_provider_lease_v1_service.ServiceStatusRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(2);
              return getClient(service).serviceStatus(input, options);
            }, { path: [2, 1] }),
            /**
             * streamServiceStatus
             */
            streamServiceStatus: withMetadata(async function streamServiceStatus(input: DeepPartial<virtengine_provider_lease_v1_service.ServiceStatusRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(2);
              return getClient(service).streamServiceStatus(input, options);
            }, { path: [2, 2] }),
            /**
             * serviceLogs
             */
            serviceLogs: withMetadata(async function serviceLogs(input: DeepPartial<virtengine_provider_lease_v1_service.ServiceLogsRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(2);
              return getClient(service).serviceLogs(input, options);
            }, { path: [2, 3] }),
            /**
             * streamServiceLogs
             */
            streamServiceLogs: withMetadata(async function streamServiceLogs(input: DeepPartial<virtengine_provider_lease_v1_service.ServiceLogsRequest>, options?: CallOptions) {
              const service = await serviceLoader.loadAt(2);
              return getClient(service).streamServiceLogs(input, options);
            }, { path: [2, 4] })
          }
        },
        v1: {
          /**
           * getStatus defines a method to query provider state
           */
          getStatus: withMetadata(async function getStatus(input: DeepPartial<google_protobuf_empty.Empty> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(3);
            return getClient(service).getStatus(input, options);
          }, { path: [3, 0] }),
          /**
           * Status defines a method to stream provider state
           */
          streamStatus: withMetadata(async function streamStatus(input: DeepPartial<google_protobuf_empty.Empty> = {}, options?: CallOptions) {
            const service = await serviceLoader.loadAt(3);
            return getClient(service).streamStatus(input, options);
          }, { path: [3, 1] })
        }
      }
    }
  };
}
