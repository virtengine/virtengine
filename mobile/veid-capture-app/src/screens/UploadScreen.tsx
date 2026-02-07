import React, { useState } from "react";
import { ActivityIndicator, StyleSheet, Text, View } from "react-native";
import { CaptureFooter } from "../components/CaptureFooter";
import { CaptureHeader } from "../components/CaptureHeader";
import { buildCapturePayload, finalizeCaptureSession } from "../core/captureSession";
import { uploadCapture } from "../services/upload/captureUploader";
import { useCaptureStore } from "../state/captureStore";

export function UploadScreen() {
  const { state, dispatch } = useCaptureStore();
  const [status, setStatus] = useState<"idle" | "uploading" | "success" | "error">("idle");
  const [error, setError] = useState<string | null>(null);

  const handleUpload = async () => {
    setStatus("uploading");
    const session = finalizeCaptureSession(state.session, "0.1.0");
    const payload = buildCapturePayload(session, "https://api.virtengine.local/veid/capture", true);
    const result = await uploadCapture(payload);
    if (result.success) {
      setStatus("success");
      dispatch({ type: "next" });
    } else {
      setStatus("error");
      setError(result.error ?? "unknown_error");
    }
  };

  return (
    <View style={styles.container}>
      <CaptureHeader
        title="Secure Upload"
        stepIndex={6}
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
