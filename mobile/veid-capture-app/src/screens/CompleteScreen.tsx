import React from "react";
import { StyleSheet, Text, View } from "react-native";

export function CompleteScreen() {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>Verification Submitted</Text>
      <Text style={styles.subtitle}>Your VEID capture package is now under review.</Text>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f9fafb",
    justifyContent: "center",
    alignItems: "center",
    paddingHorizontal: 20
  },
  title: {
    fontSize: 20,
    fontWeight: "700",
    color: "#111827"
  },
  subtitle: {
    marginTop: 8,
    color: "#6b7280",
    textAlign: "center"
  }
});
