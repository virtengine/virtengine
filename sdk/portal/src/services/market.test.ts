import { afterEach, describe, expect, it, vi } from "vitest";

import { fetchMarketListings } from "./market";

describe("fetchMarketListings", () => {
  afterEach(() => {
    vi.unstubAllGlobals();
  });

  it("normalizes live offering and provider records", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          offerings: [
            {
              provider_address: "virtengine1providerxyz",
              sequence: "101",
              name: "NVIDIA A100 Cluster",
              description: "GPU cluster",
              category: "gpu",
              prices: [{ price: { amount: "1250000", denom: "uve" } }],
              state: "active",
            },
          ],
        }),
      })
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          providers: [
            {
              address: "virtengine1providerxyz",
              display_name: "CloudCore",
            },
          ],
        }),
      });

    vi.stubGlobal("fetch", fetchMock);

    const listings = await fetchMarketListings();

    expect(fetchMock).toHaveBeenCalledTimes(2);
    expect(listings).toEqual([
      {
        id: "virtengine1providerxyz/101",
        title: "NVIDIA A100 Cluster",
        price: "1250000 uve",
        providerAddress: "virtengine1providerxyz",
        providerName: "CloudCore",
        category: "gpu",
        description: "GPU cluster",
        state: "active",
      },
    ]);
  });

  it("falls back to offering data when the provider directory is unavailable", async () => {
    const fetchMock = vi
      .fn()
      .mockResolvedValueOnce({
        ok: true,
        json: async () => ({
          offerings: [
            {
              id: { providerAddress: "virtengine1providerxyz", sequence: "202" },
              pricing: { base_price: "750000" },
            },
          ],
        }),
      })
      .mockResolvedValueOnce({
        ok: false,
        status: 404,
      })
      .mockResolvedValueOnce({
        ok: false,
        status: 404,
      });

    vi.stubGlobal("fetch", fetchMock);

    const listings = await fetchMarketListings();

    expect(listings[0]).toMatchObject({
      id: "virtengine1providerxyz/202",
      providerAddress: "virtengine1providerxyz",
      price: "750000 uve",
    });
    expect(listings[0]?.providerName).toBeUndefined();
  });
});
