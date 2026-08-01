import React from "react";
import { SafeAreaView, StyleSheet } from "react-native";
import { CaptureProvider, useCaptureStore } from "./src/state/captureStore";
import { ConsentScreen } from "./src/screens/ConsentScreen";
import { SocialMediaScreen } from "./src/screens/SocialMediaScreen";
import { DocumentCaptureScreen } from "./src/screens/DocumentCaptureScreen";
import { SelfieCaptureScreen } from "./src/screens/SelfieCaptureScreen";
import { LivenessScreen } from "./src/screens/LivenessScreen";
import { BiometricCaptureScreen } from "./src/screens/BiometricCaptureScreen";
import { ReviewScreen } from "./src/screens/ReviewScreen";
import { UploadScreen } from "./src/screens/UploadScreen";
import { CompleteScreen } from "./src/screens/CompleteScreen";
import type { CaptureUploadDependencies } from "./src/services/upload/captureUploadAttempt";
import { createId } from "./src/utils/id";

const productionUploadDependencies: CaptureUploadDependencies = {
  createIdempotencyKey: () => createId("upload")
};

function CaptureRouter({ uploadDependencies }: { uploadDependencies: CaptureUploadDependencies }) {
  const { state } = useCaptureStore();

  switch (state.currentStep) {
    case "consent":
      return <ConsentScreen />;
    case "social_media":
      return <SocialMediaScreen />;
    case "document_front":
      return <DocumentCaptureScreen side="front" stepIndex={2} />;
    case "document_back":
      return <DocumentCaptureScreen side="back" stepIndex={2} />;
    case "selfie":
      return <SelfieCaptureScreen stepIndex={3} />;
    case "liveness":
      return <LivenessScreen stepIndex={4} />;
    case "biometric":
      return <BiometricCaptureScreen />;
    case "review":
      return <ReviewScreen />;
    case "upload":
      return <UploadScreen uploadDependencies={uploadDependencies} />;
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
        <CaptureRouter uploadDependencies={productionUploadDependencies} />
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
