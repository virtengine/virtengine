import React from "react";
import { SafeAreaView, StyleSheet } from "react-native";
import { CaptureProvider, useCaptureStore } from "./src/state/captureStore";
import { ConsentScreen } from "./src/screens/ConsentScreen";
import { DocumentCaptureScreen } from "./src/screens/DocumentCaptureScreen";
import { SelfieCaptureScreen } from "./src/screens/SelfieCaptureScreen";
import { LivenessScreen } from "./src/screens/LivenessScreen";
import { ReviewScreen } from "./src/screens/ReviewScreen";
import { UploadScreen } from "./src/screens/UploadScreen";
import { CompleteScreen } from "./src/screens/CompleteScreen";

function CaptureRouter() {
  const { state } = useCaptureStore();

  switch (state.currentStep) {
    case "consent":
      return <ConsentScreen />;
    case "document_front":
      return <DocumentCaptureScreen side="front" stepIndex={1} />;
    case "document_back":
      return <DocumentCaptureScreen side="back" stepIndex={1} />;
    case "selfie":
      return <SelfieCaptureScreen />;
    case "liveness":
      return <LivenessScreen />;
    case "review":
      return <ReviewScreen />;
    case "upload":
      return <UploadScreen />;
    case "complete":
      return <CompleteScreen />;
    default:
      return <ConsentScreen />;
  }
}

export default function App() {
  return (
    <CaptureProvider>
      <SafeAreaView style={styles.container}>
        <CaptureRouter />
      </SafeAreaView>
    </CaptureProvider>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f9fafb"
  }
});
