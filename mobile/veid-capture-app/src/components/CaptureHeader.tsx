import React from "react";
import { StyleSheet, Text, View } from "react-native";
import { ProgressStepper } from "./ProgressStepper";

interface CaptureHeaderProps {
  title: string;
  stepIndex: number;
  subtitle?: string;
}

export function CaptureHeader({ title, stepIndex, subtitle }: CaptureHeaderProps) {
  return (
    <View style={styles.container}>
      <Text style={styles.title}>{title}</Text>
      {subtitle ? <Text style={styles.subtitle}>{subtitle}</Text> : null}
      <ProgressStepper activeIndex={stepIndex} />
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    paddingHorizontal: 20,
    paddingTop: 20
  },
  title: {
    fontSize: 20,
    fontWeight: "700",
    color: "#111827"
  },
  subtitle: {
    fontSize: 14,
    color: "#4b5563",
    marginTop: 6
  }
});
