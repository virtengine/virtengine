import React from "react";
import { Pressable, StyleSheet, Text, View } from "react-native";

interface CaptureFooterProps {
  primaryLabel: string;
  onPrimary: () => void;
  secondaryLabel?: string;
  onSecondary?: () => void;
  disabled?: boolean;
}

export function CaptureFooter({ primaryLabel, onPrimary, secondaryLabel, onSecondary, disabled }: CaptureFooterProps) {
  return (
    <View style={styles.container}>
      {secondaryLabel ? (
        <Pressable style={styles.secondaryButton} onPress={onSecondary}>
          <Text style={styles.secondaryText}>{secondaryLabel}</Text>
        </Pressable>
      ) : null}
      <Pressable style={[styles.primaryButton, disabled ? styles.primaryDisabled : null]} onPress={onPrimary} disabled={disabled}>
        <Text style={styles.primaryText}>{primaryLabel}</Text>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    paddingHorizontal: 20,
    paddingBottom: 24,
    flexDirection: "row",
    gap: 12
  },
  primaryButton: {
    flex: 1,
    backgroundColor: "#111827",
    paddingVertical: 14,
    borderRadius: 12,
    alignItems: "center"
  },
  primaryDisabled: {
    opacity: 0.5
  },
  primaryText: {
    color: "#f9fafb",
    fontWeight: "600"
  },
  secondaryButton: {
    flex: 1,
    backgroundColor: "#e5e7eb",
    paddingVertical: 14,
    borderRadius: 12,
    alignItems: "center"
  },
  secondaryText: {
    color: "#111827",
    fontWeight: "600"
  }
});
