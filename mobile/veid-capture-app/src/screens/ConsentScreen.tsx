import React from "react";
import { ScrollView, StyleSheet, Text, View } from "react-native";
import { CaptureFooter } from "../components/CaptureFooter";
import { CaptureHeader } from "../components/CaptureHeader";
import { useCaptureStore } from "../state/captureStore";

export function ConsentScreen() {
  const { dispatch } = useCaptureStore();

  return (
    <View style={styles.container}>
      <CaptureHeader
        title="Identity Verification Consent"
        stepIndex={0}
        subtitle="Review and accept the biometric capture consent terms."
      />
      <ScrollView style={styles.content}>
        <Text style={styles.paragraph}>
          By continuing, you consent to the capture and processing of identity documents,
          biometric images, and liveness signals for the purpose of VEID verification.
        </Text>
        <Text style={styles.paragraph}>
          You can review the Biometric Data Addendum and Privacy Policy in the VirtEngine
          documentation before submitting your verification payload.
        </Text>
      </ScrollView>
      <CaptureFooter
        primaryLabel="I Agree"
        onPrimary={() => dispatch({ type: "accept_consent" })}
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
  paragraph: {
    fontSize: 14,
    color: "#374151",
    marginBottom: 12,
    lineHeight: 20
  }
});
