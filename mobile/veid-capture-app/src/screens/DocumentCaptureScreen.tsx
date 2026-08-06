import React, { useState } from "react";
import { StyleSheet, Text, View } from "react-native";
import { CameraFrame } from "../components/CameraFrame";
import { CaptureFooter } from "../components/CaptureFooter";
import { CaptureHeader } from "../components/CaptureHeader";
import { DocumentGuidance } from "../components/DocumentGuidance";
import type { DocumentSide } from "../core/captureModels";
import type { CameraAdapter } from "../services/camera/cameraAdapter";
import { useCaptureStore } from "../state/captureStore";

interface DocumentCaptureScreenProps {
  side: DocumentSide;
  stepIndex: number;
  cameraAdapter?: CameraAdapter;
}

export function DocumentCaptureScreen({ side, stepIndex, cameraAdapter }: DocumentCaptureScreenProps) {
  const { state, dispatch } = useCaptureStore();
  const [hasCapture, setHasCapture] = useState(false);
  const [terminalError, setTerminalError] = useState<string | null>(null);

  return (
    <View style={styles.container}>
      <CaptureHeader
        title="Document Capture"
        stepIndex={stepIndex}
        subtitle={`Capture the ${side} side of your document.`}
      />
      <DocumentGuidance side={side} />
      <CameraFrame
        label={`document_${side}`}
        cameraAdapter={cameraAdapter}
        onCapture={(asset) => {
          dispatch({
            type: "set_document",
            payload: {
              type: state.session.documentType,
              side,
              image: asset,
              qualityScore: 0.82,
              warnings: []
            }
          });
          setHasCapture(true);
        }}
        onFailure={(error) => {
          setHasCapture(false);
          setTerminalError(error.code);
        }}
      />
      {terminalError ? <Text style={styles.error}>Verification stopped: {terminalError}</Text> : null}
      <CaptureFooter
        primaryLabel="Continue"
        onPrimary={() => dispatch({ type: "next" })}
        secondaryLabel="Back"
        onSecondary={() => dispatch({ type: "prev" })}
        disabled={!hasCapture || Boolean(terminalError)}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f9fafb"
  },
  error: { color: "#b91c1c", paddingHorizontal: 20, paddingBottom: 8 }
});
