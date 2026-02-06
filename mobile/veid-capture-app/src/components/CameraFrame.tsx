import React, { useEffect, useMemo, useState } from "react";
import { Image, Pressable, StyleSheet, Text, View } from "react-native";
import type { ImageAsset } from "../core/captureModels";
import { mockCameraAdapter } from "../services/camera/mockCameraAdapter";
import { visionCameraAdapter } from "../services/camera/visionCameraAdapter";

interface CameraFrameProps {
  label: string;
  onCapture: (asset: ImageAsset) => void;
}

export function CameraFrame({ label, onCapture }: CameraFrameProps) {
  const isDev = typeof __DEV__ === "undefined" ? true : __DEV__;
  const adapter = useMemo(() => (isDev ? mockCameraAdapter : visionCameraAdapter), [isDev]);
  const [permissionGranted, setPermissionGranted] = useState(false);
  const [captured, setCaptured] = useState<ImageAsset | null>(null);

  useEffect(() => {
    let isMounted = true;
    adapter.requestPermission().then((granted) => {
      if (isMounted) {
        setPermissionGranted(granted);
      }
    });
    return () => {
      isMounted = false;
    };
  }, [adapter]);

  const handleCapture = async () => {
    const asset = await adapter.capturePhoto(label);
    setCaptured(asset);
    onCapture(asset);
  };

  return (
    <View style={styles.container}>
      <View style={styles.preview}>
        {captured ? (
          <Image source={{ uri: captured.uri }} style={styles.image} />
        ) : (
          <View style={styles.placeholder}>
            <Text style={styles.placeholderText}>
              {permissionGranted ? "Camera ready" : "Camera permission required"}
            </Text>
          </View>
        )}
      </View>
      <Pressable style={styles.captureButton} onPress={handleCapture}>
        <Text style={styles.captureText}>Capture</Text>
      </Pressable>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    paddingHorizontal: 20,
    paddingVertical: 16
  },
  preview: {
    height: 320,
    borderRadius: 16,
    backgroundColor: "#111827",
    justifyContent: "center",
    alignItems: "center",
    overflow: "hidden"
  },
  placeholder: {
    alignItems: "center"
  },
  placeholderText: {
    color: "#e5e7eb"
  },
  image: {
    width: "100%",
    height: "100%",
    resizeMode: "cover"
  },
  captureButton: {
    marginTop: 16,
    backgroundColor: "#4f46e5",
    paddingVertical: 12,
    borderRadius: 12,
    alignItems: "center"
  },
  captureText: {
    color: "#ffffff",
    fontWeight: "600"
  }
});
