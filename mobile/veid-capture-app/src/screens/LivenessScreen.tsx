import React, { useEffect, useMemo, useRef, useState } from "react";
import { Pressable, StyleSheet, Text, View } from "react-native";
import { CaptureFooter } from "../components/CaptureFooter";
import { CaptureHeader } from "../components/CaptureHeader";
import { LivenessPrompt } from "../components/LivenessPrompt";
import { useCaptureStore } from "../state/captureStore";
import { LivenessEngine, createDefaultChallenges, DEFAULT_LIVENESS_CONFIG } from "../core/liveness/engine";
import type { FaceSignal, LivenessUpdate } from "../core/liveness/types";

export function LivenessScreen({ stepIndex = 4 }: { stepIndex?: number }) {
  const { dispatch } = useCaptureStore();
  const challenges = useMemo(() => createDefaultChallenges(), []);
  const engineRef = useRef(new LivenessEngine(challenges, DEFAULT_LIVENESS_CONFIG));
  const [update, setUpdate] = useState<LivenessUpdate | null>(null);
  const [completed, setCompleted] = useState(false);

  useEffect(() => {
    engineRef.current.start();
    const initial = engineRef.current.update({
      timestamp: Date.now(),
      faceConfidence: 0.9,
      yaw: 0,
      pitch: 0,
      roll: 0
    });
    setUpdate(initial);
  }, []);

  const simulateSignal = async () => {
    if (!update) {
      return;
    }

    const now = Date.now();
    const signalBase: FaceSignal = {
      timestamp: now,
      faceConfidence: 0.92,
      yaw: 0,
      pitch: 0,
      roll: 0,
      leftEyeOpenProbability: 0.9,
      rightEyeOpenProbability: 0.9,
      smileProbability: 0.1
    };

    let nextUpdate = update;

    switch (update.challengeType) {
      case "blink": {
        nextUpdate = engineRef.current.update({ ...signalBase, leftEyeOpenProbability: 0.05, rightEyeOpenProbability: 0.05 });
        nextUpdate = engineRef.current.update({ ...signalBase, timestamp: now + 200, leftEyeOpenProbability: 0.9, rightEyeOpenProbability: 0.9 });
        break;
      }
      case "turn_left": {
        nextUpdate = engineRef.current.update({ ...signalBase, yaw: -20 });
        break;
      }
      case "turn_right": {
        nextUpdate = engineRef.current.update({ ...signalBase, yaw: 20 });
        break;
      }
      case "smile": {
        nextUpdate = engineRef.current.update({ ...signalBase, smileProbability: 0.9 });
        break;
      }
      case "hold_still": {
        nextUpdate = engineRef.current.update({ ...signalBase, timestamp: now + 1500 });
        break;
      }
      default:
        break;
    }

    setUpdate(nextUpdate);
    const result = engineRef.current.getResult();
    if (result.passed && result.challenges.length === challenges.length) {
      dispatch({ type: "set_liveness", payload: result });
      setCompleted(true);
    }
  };

  return (
    <View style={styles.container}>
      <CaptureHeader
        title="Liveness Check"
        stepIndex={stepIndex}
        subtitle="Complete the active liveness challenge prompts."
      />
      {update ? (
        <LivenessPrompt instruction={update.instruction} progress={update.progress} note={update.note} />
      ) : null}
      <View style={styles.simulationCard}>
        <Text style={styles.simulationTitle}>Live detection ready</Text>
        <Text style={styles.simulationText}>
          Use the simulator button to advance through challenges when native face detection is unavailable.
        </Text>
        <Pressable style={styles.simulateButton} onPress={simulateSignal}>
          <Text style={styles.simulateText}>Simulate Challenge</Text>
        </Pressable>
      </View>
      <CaptureFooter
        primaryLabel={completed ? "Continue" : "Complete Liveness"}
        onPrimary={() => dispatch({ type: "next" })}
        secondaryLabel="Back"
        onSecondary={() => dispatch({ type: "prev" })}
        disabled={!completed}
      />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: "#f9fafb"
  },
  simulationCard: {
    marginHorizontal: 20,
    marginTop: 24,
    padding: 16,
    backgroundColor: "#ffffff",
    borderRadius: 16,
    borderWidth: 1,
    borderColor: "#e5e7eb"
  },
  simulationTitle: {
    fontSize: 16,
    fontWeight: "600",
    color: "#111827"
  },
  simulationText: {
    marginTop: 6,
    fontSize: 13,
    color: "#6b7280"
  },
  simulateButton: {
    marginTop: 12,
    paddingVertical: 10,
    borderRadius: 12,
    backgroundColor: "#111827",
    alignItems: "center"
  },
  simulateText: {
    color: "#ffffff",
    fontWeight: "600"
  }
});
