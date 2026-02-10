import React, { useMemo, useState } from "react";
import { ActivityIndicator, Pressable, ScrollView, StyleSheet, Text, View } from "react-native";
import { CaptureFooter } from "../components/CaptureFooter";
import { CaptureHeader } from "../components/CaptureHeader";
import type { SocialMediaProvider } from "../core/captureModels";
import { requestSocialProfile } from "../services/social/socialMediaService";
import { useCaptureStore } from "../state/captureStore";

const providers: { id: SocialMediaProvider; label: string; helper: string }[] = [
  { id: "google", label: "Google", helper: "Profile, email, verification" },
  { id: "facebook", label: "Facebook", helper: "Profile, friend range, verification" },
  { id: "microsoft", label: "Microsoft", helper: "Profile, org membership, verification" }
];

export function SocialMediaScreen() {
  const { state, dispatch } = useCaptureStore();
  const [loadingProvider, setLoadingProvider] = useState<SocialMediaProvider | null>(null);

  const connected = useMemo(() => {
    const map = new Map<SocialMediaProvider, boolean>();
    (state.session.socialMedia ?? []).forEach((profile) => map.set(profile.provider, true));
    return map;
  }, [state.session.socialMedia]);

  const handleConnect = async (provider: SocialMediaProvider) => {
    setLoadingProvider(provider);
    const profile = await requestSocialProfile(provider);
    dispatch({ type: "add_social", payload: profile });
    setLoadingProvider(null);
  };

  return (
    <View style={styles.container}>
      <CaptureHeader
        title="Connect Social Accounts"
        stepIndex={1}
        subtitle="Link social profiles to strengthen identity scoring."
      />
      <ScrollView style={styles.content}>
        {providers.map((provider) => {
          const isConnected = connected.get(provider.id);
          const isLoading = loadingProvider === provider.id;
          return (
            <View key={provider.id} style={styles.card}>
              <View style={styles.cardHeader}>
                <Text style={styles.providerLabel}>{provider.label}</Text>
                <Text style={[styles.badge, isConnected ? styles.badgeActive : styles.badgeIdle]}>
                  {isConnected ? "Connected" : "Not connected"}
                </Text>
              </View>
              <Text style={styles.helper}>{provider.helper}</Text>
              <Pressable
                style={[styles.connectButton, isConnected ? styles.connectDisabled : null]}
                onPress={() => handleConnect(provider.id)}
                disabled={isConnected || isLoading}
              >
                {isLoading ? <ActivityIndicator color="#111827" /> : <Text style={styles.connectText}>Connect</Text>}
              </Pressable>
            </View>
          );
        })}
      </ScrollView>
      <CaptureFooter
        primaryLabel="Continue"
        onPrimary={() => dispatch({ type: "next" })}
        secondaryLabel="Back"
        onSecondary={() => dispatch({ type: "prev" })}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f9fafb"
  },
  content: {
    paddingHorizontal: 20,
    marginTop: 12
  },
  card: {
    backgroundColor: "#ffffff",
    borderRadius: 16,
    padding: 16,
    marginBottom: 16,
    shadowColor: "#111827",
    shadowOpacity: 0.08,
    shadowOffset: { width: 0, height: 6 },
    shadowRadius: 12,
    elevation: 2
  },
  cardHeader: {
    flexDirection: "row",
    justifyContent: "space-between",
    alignItems: "center"
  },
  providerLabel: {
    fontSize: 16,
    fontWeight: "600",
    color: "#111827"
  },
  badge: {
    fontSize: 12,
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: 999
  },
  badgeActive: {
    backgroundColor: "#dcfce7",
    color: "#15803d"
  },
  badgeIdle: {
    backgroundColor: "#e5e7eb",
    color: "#6b7280"
  },
  helper: {
    marginTop: 8,
    color: "#6b7280",
    fontSize: 13
  },
  connectButton: {
    marginTop: 12,
    backgroundColor: "#f3f4f6",
    borderRadius: 12,
    paddingVertical: 10,
    alignItems: "center"
  },
  connectDisabled: {
    opacity: 0.5
  },
  connectText: {
    color: "#111827",
    fontWeight: "600"
  }
});
