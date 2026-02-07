import React, { useState } from "react";
import { StyleSheet, Text, View } from "react-native";
import { CameraFrame } from "../components/CameraFrame";
import { CaptureFooter } from "../components/CaptureFooter";
import { CaptureHeader } from "../components/CaptureHeader";
import { useCaptureStore } from "../state/captureStore";

export function SelfieCaptureScreen({ stepIndex = 2 }: { stepIndex?: number }) {
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
