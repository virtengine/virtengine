import { useMemo } from "react";
import {
  ProviderAPIClient,
  type ProviderAPIClientOptions,
} from "../provider-api/client";

/**
 * React hook that memoizes a {@link ProviderAPIClient} instance.
 *
 * A new client is created only when the endpoint, authentication config, or
 * fetcher reference changes.
 */
export function useProviderAPI(
  options: ProviderAPIClientOptions,
): ProviderAPIClient {
  const walletAddress = options.wallet?.address;
  const walletChainId = options.wallet?.chainId;
  const hmacSecret = options.hmac?.secret;
  const hmacPrincipal = options.hmac?.principal;

  return useMemo(
    () => new ProviderAPIClient(options),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [
      options.endpoint,
      walletAddress,
      walletChainId,
      hmacSecret,
      hmacPrincipal,
      options.fetcher,
    ],
  );
}
