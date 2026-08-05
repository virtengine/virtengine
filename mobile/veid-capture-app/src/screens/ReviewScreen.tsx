import React, { useEffect, useState } from "react";
import { ActivityIndicator, ScrollView, StyleSheet, Text, View } from "react-native";
import { CaptureFooter } from "../components/CaptureFooter";
import { CaptureHeader } from "../components/CaptureHeader";
import { extractOcr } from "../services/ocr/ocrService";
import { useCaptureStore } from "../state/captureStore";
import { VerificationTerminalError } from "../core/verificationError";
import { hasCalibratedFieldConfidence } from "../core/ocr/confidence";

export function ReviewScreen() {
  const { state, dispatch } = useCaptureStore();
  const [loading, setLoading] = useState(false);
  const [terminalError, setTerminalError] = useState<string | null>(null);
  const biometricStatus = state.session.biometric
    ? state.session.biometric.supported
      ? "Captured"
      : "Unsupported"
    : "Pending";
  const attestationStatus = state.session.deviceAttestation
    ? state.session.deviceAttestation.supported
      ? "Verified"
      : "Unsupported"
    : "Pending";
  const socialStatus = state.session.socialMedia?.length
    ? `${state.session.socialMedia.length} connected`
    : "None";

  useEffect(() => {
    const runOcr = async () => {
      if (!state.session.documentFront || state.session.ocr) {
        return;
      }
      setLoading(true);
      try {
        const result = await extractOcr(state.session.documentFront.image.uri);
        if (!hasCalibratedFieldConfidence(result)) {
          throw new VerificationTerminalError("ocr_confidence_unavailable", "OCR output lacks calibrated field confidence and cannot be used for verification.");
        }
        dispatch({ type: "set_ocr", payload: result });
      } catch (error) {
        setTerminalError(error instanceof VerificationTerminalError ? error.code : "ocr_recognition_failed");
      } finally {
        setLoading(false);
      }
    };

    runOcr();
  }, [state.session.documentFront, state.session.ocr, dispatch]);

  return (
    <View style={styles.container}>
      <CaptureHeader
        title="Review Capture"
        stepIndex={6}
        subtitle="Confirm your document, biometric, and liveness status."
      />
      <ScrollView style={styles.content}>
        <Text style={styles.sectionTitle}>Captured Assets</Text>
        <Text style={styles.line}>Document front: {state.session.documentFront ? "Ready" : "Missing"}</Text>
        <Text style={styles.line}>Document back: {state.session.documentBack ? "Ready" : "Missing"}</Text>
        <Text style={styles.line}>Selfie: {state.session.selfie ? "Ready" : "Missing"}</Text>
        <Text style={styles.line}>Liveness: {state.session.liveness?.passed ? "Passed" : "Pending"}</Text>
        <Text style={styles.line}>Biometric hardware: {biometricStatus}</Text>
        <Text style={styles.line}>Device attestation: {attestationStatus}</Text>
        <Text style={styles.line}>Social accounts: {socialStatus}</Text>

        <Text style={styles.sectionTitle}>OCR Fields</Text>
        {loading ? <ActivityIndicator /> : null}
        {terminalError ? <Text style={styles.error}>Verification stopped: {terminalError}</Text> : null}
        {state.session.ocr?.fields.map((field) => (
          <Text key={field.key} style={styles.line}>
            {field.key}: {field.value}
          </Text>
        ))}
      </ScrollView>
      <CaptureFooter
        primaryLabel="Submit"
        onPrimary={() => dispatch({ type: "next" })}
        secondaryLabel="Back"
        onSecondary={() => dispatch({ type: "prev" })}
        disabled={!state.session.liveness?.passed || !state.session.ocr || Boolean(terminalError)}
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
  sectionTitle: {
    marginTop: 16,
    fontSize: 14,
    fontWeight: "600",
    color: "#111827"
  },
  line: {
    marginTop: 8,
    color: "#4b5563",
    fontSize: 13
  },
  error: { marginTop: 8, color: "#b91c1c", fontSize: 13, fontWeight: "600" }
});
