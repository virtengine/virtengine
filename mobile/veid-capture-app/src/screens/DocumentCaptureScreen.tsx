import React, { useState } from "react";
import { StyleSheet, View } from "react-native";
import { CameraFrame } from "../components/CameraFrame";
import { CaptureFooter } from "../components/CaptureFooter";
import { CaptureHeader } from "../components/CaptureHeader";
import { DocumentGuidance } from "../components/DocumentGuidance";
import type { DocumentSide } from "../core/captureModels";
import { useCaptureStore } from "../state/captureStore";

interface DocumentCaptureScreenProps {
  side: DocumentSide;
  stepIndex: number;
}

export function DocumentCaptureScreen({ side, stepIndex }: DocumentCaptureScreenProps) {
  const { state, dispatch } = useCaptureStore();
  const [hasCapture, setHasCapture] = useState(false);

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
      />
      <CaptureFooter
        primaryLabel="Continue"
        onPrimary={() => dispatch({ type: "next" })}
        secondaryLabel="Back"
        onSecondary={() => dispatch({ type: "prev" })}
        disabled={!hasCapture}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f9fafb"
  }
});
