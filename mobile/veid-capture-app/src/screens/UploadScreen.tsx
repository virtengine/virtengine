import React, { useState } from "react";
import { ActivityIndicator, StyleSheet, Text, View } from "react-native";
import { CaptureFooter } from "../components/CaptureFooter";
import { CaptureHeader } from "../components/CaptureHeader";
import { buildCapturePayload, finalizeCaptureSession } from "../core/captureSession";
import type { DeviceAttestationProviderAdapter } from "../core/deviceAttestation";
import { uploadCapture } from "../services/upload/captureUploader";
import { useCaptureStore } from "../state/captureStore";

interface UploadScreenProps {
  attestationProvider?: DeviceAttestationProviderAdapter;
}

export function UploadScreen({ attestationProvider }: UploadScreenProps) {
  const { state, dispatch } = useCaptureStore();
  const [status, setStatus] = useState<"idle" | "uploading" | "success" | "error">("idle");
  const [error, setError] = useState<string | null>(null);

  const handleUpload = async () => {
    setStatus("uploading");
    setError(null);

    try {
      const session = finalizeCaptureSession(state.session, "0.1.0", attestationProvider);
      if (!session.deviceAttestation?.supported) {
        setStatus("error");
        setError(session.deviceAttestation?.failureReason ?? "attestation_unavailable");
        return;
      }

      const payload = buildCapturePayload(session, "https://api.virtengine.local/veid/capture");
      const result = await uploadCapture(payload);
      if (result.success) {
        setStatus("success");
        dispatch({ type: "next" });
        return;
      }

      setStatus("error");
      setError(result.error ?? "unknown_error");
    } catch (uploadError) {
      setStatus("error");
      setError(uploadError instanceof Error ? uploadError.message : "upload_unavailable");
    }
  };

  return (
    <View style={styles.container}>
      <CaptureHeader
        title="Secure Upload"
        stepIndex={7}
        subtitle="Encrypt and transmit your capture package with attestation."
      />
      <View style={styles.content}>
        {status === "uploading" ? <ActivityIndicator /> : null}
        {status === "success" ? <Text style={styles.success}>Upload complete.</Text> : null}
        {status === "error" ? <Text style={styles.error}>Upload failed: {error}</Text> : null}
      </View>
      <CaptureFooter
        primaryLabel={status === "success" ? "Finish" : "Upload"}
        onPrimary={handleUpload}
        secondaryLabel="Back"
        onSecondary={() => dispatch({ type: "prev" })}
        disabled={status === "uploading"}
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
    flex: 1,
    paddingHorizontal: 20,
    justifyContent: "center",
    alignItems: "center"
  },
  success: {
    color: "#16a34a",
    fontWeight: "600"
  },
  error: {
    color: "#dc2626",
    fontWeight: "600"
  }
});
