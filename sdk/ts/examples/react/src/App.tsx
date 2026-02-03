import { useMemo, useState } from "react";
import { createChainNodeWebSDK } from "@virtengine/chain-sdk";

const defaultGateway = "https://api.sandbox-2.aksh.pw:443";

export function App() {
  const sdk = useMemo(() => createChainNodeWebSDK({
    query: {
      baseUrl: import.meta.env.VITE_VE_GRPC_GATEWAY ?? defaultGateway,
    },
  }), []);

  const [status, setStatus] = useState("idle");
  const [output, setOutput] = useState<string>("");

  const fetchDeployments = async () => {
    setStatus("loading");
    try {
      const result = await sdk.virtengine.deployment.v1beta4.getDeployments({
        pagination: { limit: 1 },
      });
      setOutput(JSON.stringify(result, null, 2));
      setStatus("success");
    } catch (error) {
      setOutput(String(error));
      setStatus("error");
    }
  };

  return (
    <div style={{ fontFamily: "system-ui", padding: "2rem", maxWidth: 720 }}>
      <h1>VirtEngine SDK + React</h1>
      <p>Query the latest deployment from the chain.</p>
      <button onClick={fetchDeployments} disabled={status === "loading"}>
        {status === "loading" ? "Loading..." : "Fetch deployments"}
      </button>
      <pre style={{ marginTop: "1.5rem", background: "#111", color: "#fafafa", padding: "1rem", borderRadius: 8 }}>
        {output || "No data yet."}
      </pre>
    </div>
  );
}
