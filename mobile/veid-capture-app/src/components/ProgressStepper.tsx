import React from "react";
import { StyleSheet, Text, View } from "react-native";

const steps = ["Consent", "Document", "Selfie", "Liveness", "Review", "Upload"];

export function ProgressStepper({ activeIndex }: { activeIndex: number }) {
  return (
    <View style={styles.container}>
      {steps.map((label, index) => (
        <View key={label} style={styles.step}>
          <View style={[styles.dot, index <= activeIndex ? styles.dotActive : null]} />
          <Text style={[styles.label, index <= activeIndex ? styles.labelActive : null]}>{label}</Text>
        </View>
      ))}
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flexDirection: "row",
    justifyContent: "space-between",
    paddingVertical: 8
  },
  step: {
    alignItems: "center",
    flex: 1
  },
  dot: {
    width: 10,
    height: 10,
    borderRadius: 5,
    backgroundColor: "#2e2e2e",
    marginBottom: 4
  },
  dotActive: {
    backgroundColor: "#4ade80"
  },
  label: {
    fontSize: 10,
    color: "#6b7280"
  },
  labelActive: {
    color: "#111827",
    fontWeight: "600"
  }
});
