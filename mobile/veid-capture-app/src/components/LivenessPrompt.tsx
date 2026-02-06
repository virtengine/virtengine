import React from "react";
import { StyleSheet, Text, View } from "react-native";

interface LivenessPromptProps {
  instruction: string;
  progress: number;
  note?: string;
}

export function LivenessPrompt({ instruction, progress, note }: LivenessPromptProps) {
  return (
    <View style={styles.container}>
      <Text style={styles.instruction}>{instruction}</Text>
      {note ? <Text style={styles.note}>{note}</Text> : null}
      <View style={styles.progressTrack}>
        <View style={[styles.progressFill, { width: `${Math.min(progress, 1) * 100}%` }]} />
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    paddingHorizontal: 20,
    paddingVertical: 16
  },
  instruction: {
    fontSize: 16,
    fontWeight: "600",
    color: "#111827"
  },
  note: {
    marginTop: 6,
    color: "#6b7280",
    fontSize: 12
  },
  progressTrack: {
    marginTop: 12,
    height: 6,
    backgroundColor: "#e5e7eb",
    borderRadius: 4,
    overflow: "hidden"
  },
  progressFill: {
    height: "100%",
    backgroundColor: "#4ade80"
  }
});
