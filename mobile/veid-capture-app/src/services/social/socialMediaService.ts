import type { SocialMediaProfile, SocialMediaProvider } from "../../core/captureModels";
import { hashString } from "../../utils/hash";

const providerSeeds: Record<SocialMediaProvider, { ageDays: number; verified: boolean; friendRange?: string }> = {
  google: { ageDays: 1460, verified: true },
  facebook: { ageDays: 3650, verified: false, friendRange: "200-500" },
  microsoft: { ageDays: 2200, verified: true }
};

export async function requestSocialProfile(provider: SocialMediaProvider): Promise<SocialMediaProfile> {
  const seed = providerSeeds[provider];
  const now = Date.now();

  return {
    provider,
    profileNameHash: hashString(`${provider}_profile_name`),
    emailHash: provider === "facebook" ? undefined : hashString(`${provider}_email@example.com`),
    usernameHash: provider === "facebook" ? hashString(`${provider}_handle`) : undefined,
    orgHash: provider === "microsoft" ? hashString("contoso") : undefined,
    accountAgeDays: seed.ageDays,
    accountCreatedAt: now - seed.ageDays * 24 * 60 * 60 * 1000,
    isVerified: seed.verified,
    friendCountRange: seed.friendRange,
    connectedAt: now
  };
}
