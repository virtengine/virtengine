import React, { useRef, useState } from "react";
import { ActivityIndicator, StyleSheet, Text, View } from "react-native";
import { CaptureFooter } from "../components/CaptureFooter";
import { CaptureHeader } from "../components/CaptureHeader";
import { buildCapturePayload, finalizeCaptureSession } from "../core/captureSession";
import type { DeviceAttestationProviderAdapter } from "../core/deviceAttestation";
import {
  createCaptureUploadAttempt,
  type CaptureUploadAttempt,
  type CaptureUploadDependencies
} from "../services/upload/captureUploadAttempt";
import {
  createCaptureCleanupCoordinator,
  type CaptureCleanupDependencies
} from "../services/cleanup/captureCleanup";
import { useCaptureStore } from "../state/captureStore";

export interface UploadScreenProps {
  attestationProvider?: DeviceAttestationProviderAdapter;
  uploadDependencies: CaptureUploadDependencies;
  cleanupDependencies?: CaptureCleanupDependencies;
}

export function UploadScreen({
  attestationProvider,
  uploadDependencies,
  cleanupDependencies = {}
}: UploadScreenProps) {
  const { state, dispatch } = useCaptureStore();
  const [status, setStatus] = useState<"idle" | "uploading" | "success" | "error">("idle");
  const [error, setError] = useState<string | null>(null);
  const uploadAttempt = useRef<CaptureUploadAttempt>();
  const pendingAcknowledgement = useRef<{ acknowledgement: unknown; cleanupId: string }>();
  const cleanupCoordinator = useRef(createCaptureCleanupCoordinator(cleanupDependencies));

  const artifactUris = () =>
    [
      state.session.documentFront?.image.uri,
      state.session.documentBack?.image.uri,
      state.session.selfie?.image.uri
    ].filter((uri): uri is string => Boolean(uri));

  const wipeSensitiveData = () => {
    uploadAttempt.current?.wipe();
    uploadAttempt.current = undefined;
    dispatch({ type: "wipe" });
  };

  const finishCleanup = async (acknowledgement: unknown, cleanupId: string, uris: string[]) => {
    const cleanup = await cleanupCoordinator.current.afterDurableAcknowledgement(acknowledgement, {
      cleanupId,
      artifactUris: uris,
      wipeSensitiveData
    });
    if (!cleanup.success) {
      pendingAcknowledgement.current = { acknowledgement, cleanupId };
      setStatus("error");
      setError(cleanup.error);
      return;
    }

    pendingAcknowledgement.current = undefined;
    setStatus("success");
    dispatch({ type: "complete" });
  };

  const handleUpload = async () => {
    setStatus("uploading");
    setError(null);

    try {
      if (pendingAcknowledgement.current) {
        await finishCleanup(
          pendingAcknowledgement.current.acknowledgement,
          pendingAcknowledgement.current.cleanupId,
          []
        );
        return;
      }

      const session = finalizeCaptureSession(state.session, "0.1.0", attestationProvider);
      if (!session.deviceAttestation?.supported) {
        setStatus("error");
        setError(session.deviceAttestation?.failureReason ?? "attestation_unavailable");
        return;
      }

      if (!uploadAttempt.current) {
        const payload = buildCapturePayload(session, "https://api.virtengine.local/veid/capture");
        uploadAttempt.current = createCaptureUploadAttempt(payload, uploadDependencies);
      }

      const result = await uploadAttempt.current.upload();
      if (result.success) {
        await finishCleanup(result.receipt, uploadAttempt.current.idempotencyKey, artifactUris());
        return;
      }

      setStatus("error");
      setError(result.error);
    } catch (uploadError) {
      setStatus("error");
      setError(uploadError instanceof Error ? uploadError.message : "upload_unavailable");
    }
  };

  const handleCancel = async () => {
    setStatus("uploading");
    setError(null);
    const cleanupId = uploadAttempt.current?.idempotencyKey ?? uploadDependencies.createIdempotencyKey();
    const result = await cleanupCoordinator.current.cancel({
      cleanupId,
      artifactUris: artifactUris(),
      wipeSensitiveData
    });
    if (!result.success) {
      setStatus("error");
      setError(result.error);
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
        secondaryLabel="Cancel"
        onSecondary={handleCancel}
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
