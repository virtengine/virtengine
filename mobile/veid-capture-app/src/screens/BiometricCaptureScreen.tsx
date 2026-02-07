import React, { useState } from "react";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { CaptureFooter } from "../components/CaptureFooter";
import { CaptureHeader } from "../components/CaptureHeader";
import { captureBiometric } from "../core/biometric";
import type { BiometricCapture, BiometricModality } from "../core/captureModels";
import { useCaptureStore } from "../state/captureStore";

export function BiometricCaptureScreen() {
  const { state, dispatch } = useCaptureStore();
  const [modality, setModality] = useState<BiometricModality>("fingerprint");
  const [capture, setCapture] = useState<BiometricCapture | undefined>(state.session.biometric);

  const handleCapture = () => {
    const result = captureBiometric(modality);
    dispatch({ type: "set_biometric", payload: result });
    setCapture(result);
  };

  return (
    <View style={styles.container}>
      <CaptureHeader
        title="Biometric Hardware"
        stepIndex={4}
        subtitle="Capture fingerprint or iris using the secure sensor."
      />
      <View style={styles.selectorRow}>
        {(["fingerprint", "iris"] as BiometricModality[]).map((option) => (
          <Pressable
            key={option}
            style={[styles.selector, modality === option ? styles.selectorActive : null]}
            onPress={() => setModality(option)}
          >
            <Text style={[styles.selectorText, modality === option ? styles.selectorTextActive : null]}>
              {option === "fingerprint" ? "Fingerprint" : "Iris"}
            </Text>
          </Pressable>
        ))}
      </View>
      <View style={styles.card}>
        <Text style={styles.cardTitle}>Sensor Status</Text>
        <Text style={styles.cardLine}>
          Capture: {capture ? (capture.supported ? "Ready" : "Unsupported") : "Pending"}
        </Text>
        {capture ? (
          <>
            <Text style={styles.cardLine}>Liveness: {capture.liveness.passed ? "Passed" : "Failed"}</Text>
            <Text style={styles.cardLine}>Anti-spoof: {capture.antiSpoofing.passed ? "Passed" : "Failed"}</Text>
            <Text style={styles.cardLine}>Security: {capture.deviceInfo.securityLevel}</Text>
          </>
        ) : null}
        <Pressable style={styles.captureButton} onPress={handleCapture}>
          <Text style={styles.captureText}>Capture {modality === "fingerprint" ? "Fingerprint" : "Iris"}</Text>
        </Pressable>
      </View>
      <CaptureFooter
        primaryLabel="Continue"
        onPrimary={() => dispatch({ type: "next" })}
        secondaryLabel="Back"
        onSecondary={() => dispatch({ type: "prev" })}
        disabled={!capture}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f9fafb"
  },
  selectorRow: {
    flexDirection: "row",
    paddingHorizontal: 20,
    marginTop: 12,
    gap: 12
  },
  selector: {
    flex: 1,
    borderRadius: 12,
    borderWidth: 1,
    borderColor: "#d1d5db",
    paddingVertical: 10,
    alignItems: "center"
  },
  selectorActive: {
    backgroundColor: "#111827",
    borderColor: "#111827"
  },
  selectorText: {
    color: "#111827",
    fontWeight: "600"
  },
  selectorTextActive: {
    color: "#ffffff"
  },
  card: {
    marginTop: 20,
    marginHorizontal: 20,
    padding: 16,
    backgroundColor: "#ffffff",
    borderRadius: 16,
    borderWidth: 1,
    borderColor: "#e5e7eb"
  },
  cardTitle: {
    fontSize: 16,
    fontWeight: "600",
    color: "#111827"
  },
  cardLine: {
    marginTop: 6,
    fontSize: 13,
    color: "#4b5563"
  },
  captureButton: {
    marginTop: 14,
    paddingVertical: 12,
    borderRadius: 12,
    backgroundColor: "#111827",
    alignItems: "center"
  },
  captureText: {
    color: "#ffffff",
    fontWeight: "600"
  }
});
