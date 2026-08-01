import React, { useState } from "react";
import { StyleSheet, Text, View } from "react-native";
import { CameraFrame } from "../components/CameraFrame";
import { CaptureFooter } from "../components/CaptureFooter";
import { CaptureHeader } from "../components/CaptureHeader";
import type { CameraAdapter } from "../services/camera/cameraAdapter";
import { useCaptureStore } from "../state/captureStore";

interface SelfieCaptureScreenProps {
  stepIndex?: number;
  cameraAdapter?: CameraAdapter;
}

export function SelfieCaptureScreen({ stepIndex = 3, cameraAdapter }: SelfieCaptureScreenProps) {
  const { dispatch } = useCaptureStore();
  const [hasCapture, setHasCapture] = useState(false);

  return (
    <View style={styles.container}>
      <CaptureHeader
        title="Selfie Capture"
        stepIndex={stepIndex}
        subtitle="Ensure your face is centered and well-lit."
      />
      <Text style={styles.guidance}>Remove glasses and keep a neutral expression.</Text>
      <CameraFrame
        label="selfie"
        cameraAdapter={cameraAdapter}
        onCapture={(asset) => {
          dispatch({
            type: "set_selfie",
            payload: {
              image: asset,
              faceConfidence: 0.9,
              guidance: []
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
  },
  guidance: {
    paddingHorizontal: 20,
    paddingTop: 12,
    color: "#6b7280",
    fontSize: 13
  }
});
