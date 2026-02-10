import { Empty } from "../../../google/protobuf/empty.ts";
import { Node } from "./node.ts";
import { Cluster } from "./cluster.ts";

export const NodeRPC = {
  typeName: "virtengine.inventory.v1.NodeRPC",
  methods: {
    queryNode: {
      name: "QueryNode",
      httpPath: "/v1/node",
      input: Empty,
      output: Node,
      get parent() { return NodeRPC; },
    },
    streamNode: {
      name: "StreamNode",
      kind: "server_streaming",
      input: Empty,
      output: Node,
      get parent() { return NodeRPC; },
    },
  },
} as const;
export const ClusterRPC = {
  typeName: "virtengine.inventory.v1.ClusterRPC",
  methods: {
    queryCluster: {
      name: "QueryCluster",
      httpPath: "/v1/inventory",
      input: Empty,
      output: Cluster,
      get parent() { return ClusterRPC; },
    },
    streamCluster: {
      name: "StreamCluster",
      kind: "server_streaming",
      input: Empty,
      output: Cluster,
      get parent() { return ClusterRPC; },
    },
  },
} as const;
