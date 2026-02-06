import React from "react";
import { StyleSheet, Text, View } from "react-native";

export function DocumentGuidance({ side }: { side: "front" | "back" }) {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Capture the {side} of your document</Text>
      <Text style={styles.subtitle}>Ensure all corners are visible and text is readable.</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    paddingHorizontal: 20,
    paddingTop: 16
  },
  title: {
    fontSize: 16,
    fontWeight: "600",
    color: "#111827"
  },
  subtitle: {
    fontSize: 13,
    color: "#6b7280",
    marginTop: 4
  }
});
