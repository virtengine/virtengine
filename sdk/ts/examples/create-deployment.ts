import { createChainNodeSDK, createStargateClient, SDL, type TxInput, type QueryInput } from "@virtengine/chain-sdk";
import { DirectSecp256k1HdWallet } from "@cosmjs/proto-signing";
import type { MsgCreateDeployment } from "@virtengine/chain-sdk/private-types/virtengine.v1beta4";
import type { MsgCreateLease } from "@virtengine/chain-sdk/private-types/virtengine.v1beta5";
import { type DeploymentID, Source } from "@virtengine/chain-sdk/private-types/virtengine.v1";
import { setTimeout as wait } from "node:timers/promises";


const testMnemonic = process.env.MNEMONIC;
if (!testMnemonic) {
  throw new Error("MNEMONIC environment variable is required");
}

const wallet = await DirectSecp256k1HdWallet.fromMnemonic(testMnemonic, { prefix: "virt" });
const [account] = await wallet.getAccounts();

console.log(`Test Account Address: ${account.address}`);

// Defaults to sandbox-2.aksh.pw
const QUERY_GRPC_URL = process.env.QUERY_GRPC_URL || "http://grpc.sandbox-2.aksh.pw:9090";
const TX_RPC_URL = process.env.TX_RPC_URL || "https://rpc.sandbox-2.aksh.pw:443";

const sdk = createChainNodeSDK({
  query: {
    baseUrl: QUERY_GRPC_URL,
  },
  tx: {
    signer: createStargateClient({
      baseUrl: TX_RPC_URL,
      signer: wallet,
    })
  },
});

console.log("Step 1: Creating deployment...");
const sdl = SDL.fromString(`
# Welcome to the VirtEngine! 🚀☁
# This file is called a Stack Definition Laguage (SDL)
# SDL is a human friendly data standard for declaring deployment attributes.
# The SDL file is a "form" to request resources from the Network.
# SDL is compatible with the YAML standard and similar to Docker Compose files.

---
# Indicates version of VirtEngine configuration file. Currently only "2.0" is accepted.
version: "2.0"

# The top-level services entry contains a map of workloads to be ran on the VirtEngine deployment. Each key is a service name; values are a map containing the following keys:
# https://virtengine.network/docs/getting-started/stack-definition-language/#services
services:
  # The name of the service "web"
  web:
    # The docker container image with version. You must specify a version, the "latest" tag doesn't work.
    image: virtengine/hello-world:1.0.0
    # You can map ports here https://virtengine.network/docs/getting-started/stack-definition-language/#servicesexpose
    expose:
      - port: 3000
        as: 80
        to:
          - global: true

# The profiles section contains named compute and placement profiles to be used in the deployment.
# https://virtengine.network/docs/getting-started/stack-definition-language/#profiles
profiles:
  # profiles.compute is map of named compute profiles. Each profile specifies compute resources to be leased for each service instance uses uses the profile.
  # https://virtengine.network/docs/getting-started/stack-definition-language/#profilescompute
  compute:
    # The name of the service
    web:
      resources:
        cpu:
          units: 0.5
        memory:
          size: 512Mi
        storage:
          size: 512Mi

# profiles.placement is map of named datacenter profiles. Each profile specifies required datacenter attributes and pricing configuration for each compute profile that will be used within the datacenter. It also specifies optional list of signatures of which tenants expects audit of datacenter attributes.
# https://virtengine.network/docs/getting-started/stack-definition-language/#profilesplacement
  placement:
    dcloud:
      pricing:
        # The name of the service
        web:
          denom: uakt
          amount: 10000

# The deployment section defines how to deploy the services. It is a mapping of service name to deployment configuration.
# https://virtengine.network/docs/getting-started/stack-definition-language/#deployment
deployment:
  # The name of the service
  web:
    dcloud:
      profile: web
      count: 1
`, "beta3");

const latestBlockResponse = await sdk.cosmos.base.tendermint.v1beta1.getLatestBlock();
const deploymentMessage: TxInput<MsgCreateDeployment> = {
  id: {
    owner: account.address,
    dseq: latestBlockResponse.block?.header?.height!,
  },
  groups: sdl.groups(),
  hash: await sdl.manifestVersion(),
  deposit: {
    amount: {
      denom: "uakt",
      amount: "500000",
    },
    sources: [Source.balance],
  },
};

console.log(`Creating deployment with dseq: ${latestBlockResponse.block?.header?.height}`);
await sdk.virtengine.deployment.v1beta4.createDeployment(deploymentMessage, {
  memo: "Test deployment for lease creation - VirtEngine Chain SDK",
});

console.log("Deployment created successfully!");

const deploymentId: QueryInput<DeploymentID> = {
  owner: account.address,
  dseq: deploymentMessage.id!.dseq,
};

console.log("Step 2: Waiting for providers to create bids...");
console.log(`Deployment ID: ${deploymentId.owner}/${deploymentId.dseq}`);
let bidsResponse;
let attempts = 0;
const maxAttempts = 18;

do {
  await wait(10000);
  attempts++;

  console.log(`Checking for bids (attempt ${attempts}/${maxAttempts})...`);
  console.log("Make sure your address is whitelisted on this network.");

  bidsResponse = await sdk.virtengine.market.v1beta5.getBids({
    filters: {
      owner: deploymentId.owner,
      dseq: deploymentId.dseq,
      gseq: 1,
      oseq: 1,
    },
  });

  console.log(`Found ${bidsResponse?.bids?.length || 0} bids`);
} while ((!bidsResponse?.bids || bidsResponse.bids.length === 0) && attempts < maxAttempts);

if (bidsResponse?.bids?.length > 0) {
  console.log(`Found ${bidsResponse!.bids!.length} bids for the deployment`);
  bidsResponse?.bids?.forEach((bidResponse, index) => {
    const bid = bidResponse.bid;
    console.log(`  Bid ${index + 1}: Provider ${bid?.id?.provider}, Price: ${bid?.price?.amount}${bid?.price?.denom}`);
  });
} else {
  throw new Error(`No bids found after ${maxAttempts} attempts. Check deployment resources and pricing.`);
}

console.log("Step 4: Selecting the first bid...");
const firstBid = bidsResponse!.bids![0]!.bid!;

console.log(`Selected bid from provider: ${firstBid.id!.provider}`);

console.log("Step 5: Creating lease from selected bid...");
const leaseMessage: TxInput<MsgCreateLease> = {
  bidId: {
    owner: firstBid.id!.owner,
    dseq: firstBid.id!.dseq,
    gseq: firstBid.id!.gseq,
    oseq: firstBid.id!.oseq,
    provider: firstBid.id!.provider,
    bseq: firstBid.id!.bseq,
  },
};

await sdk.virtengine.market.v1beta5.createLease(leaseMessage, {
  memo: "Test lease creation from bid - VirtEngine Chain SDK",
});

console.log("Step 6: Verifying lease creation...");
console.log("Lease created successfully!");

const leaseQuery = await sdk.virtengine.market.v1beta5.getLeases({
  filters: {
    owner: deploymentId.owner,
    dseq: deploymentId.dseq,
    gseq: 1,
    oseq: 1,
    provider: firstBid.id!.provider,
    state: "",
    bseq: 0,
  },
});

const createdLease = leaseQuery!.leases![0]!.lease!;
console.log("Lease verification completed successfully!");
console.log(`Lease ID: ${createdLease.id?.owner}/${createdLease.id?.dseq}/${createdLease.id?.gseq}/${createdLease.id?.oseq}/${createdLease.id?.provider}`);
console.log(`Lease State: ${createdLease.state}`);
console.log(`Lease Price: ${createdLease.price?.amount}${createdLease.price?.denom}`);
