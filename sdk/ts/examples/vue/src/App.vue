<script setup lang="ts">
import { ref } from "vue";
import { createChainNodeWebSDK } from "@virtengine/chain-sdk";

const defaultGateway = "https://api.sandbox-2.aksh.pw:443";
const sdk = createChainNodeWebSDK({
  query: { baseUrl: import.meta.env.VITE_VE_GRPC_GATEWAY ?? defaultGateway },
});

const status = ref("idle");
const output = ref("No data yet.");

const fetchDeployments = async () => {
  status.value = "loading";
  try {
    const result = await sdk.virtengine.deployment.v1beta4.getDeployments({
      pagination: { limit: 1 },
    });
    output.value = JSON.stringify(result, null, 2);
    status.value = "success";
  } catch (error) {
    output.value = String(error);
    status.value = "error";
  }
};
</script>

<template>
  <div style="font-family: system-ui; padding: 2rem; max-width: 720px;">
    <h1>VirtEngine SDK + Vue</h1>
    <p>Query the latest deployment from the chain.</p>
    <button :disabled="status === 'loading'" @click="fetchDeployments">
      {{ status === "loading" ? "Loading..." : "Fetch deployments" }}
    </button>
    <pre style="margin-top: 1.5rem; background: #111; color: #fafafa; padding: 1rem; border-radius: 8px;">
      {{ output }}
    </pre>
  </div>
</template>
