import React, { useEffect, useState } from "react";
import { Image, Pressable, StyleSheet, Text, View } from "react-native";
import type { ImageAsset } from "../core/captureModels";
import type { CameraAdapter } from "../services/camera/cameraAdapter";
import { CameraUnavailableError } from "../services/camera/cameraAdapter";
import { visionCameraAdapter } from "../services/camera/visionCameraAdapter";

interface CameraFrameProps {
  label: string;
  onCapture: (asset: ImageAsset) => void;
  cameraAdapter?: CameraAdapter;
}

export function CameraFrame({ label, onCapture, cameraAdapter = visionCameraAdapter }: CameraFrameProps) {
  const [permissionGranted, setPermissionGranted] = useState(false);
  const [captured, setCaptured] = useState<ImageAsset | null>(null);
  const [unavailableReason, setUnavailableReason] = useState<string | null>(null);

  useEffect(() => {
    let isMounted = true;
    cameraAdapter.requestPermission()
      .then((granted) => {
        if (isMounted) {
          setPermissionGranted(granted);
          setUnavailableReason(granted ? null : "camera_permission_denied");
        }
      })
      .catch((error: unknown) => {
        if (isMounted) {
          setPermissionGranted(false);
          setUnavailableReason(getCameraFailureReason(error));
        }
      });
    return () => {
      isMounted = false;
    };
  }, [cameraAdapter]);

  const handleCapture = async () => {
    try {
      const asset = await cameraAdapter.capturePhoto(label);
      setCaptured(asset);
      setUnavailableReason(null);
      onCapture(asset);
    } catch (error) {
      setUnavailableReason(getCameraFailureReason(error));
    }
  };

  return (
    <View style={styles.container}>
      <View style={styles.preview}>
        {captured ? (
          <Image source={{ uri: captured.uri }} style={styles.image} />
        ) : (
          <View style={styles.placeholder}>
            <Text style={styles.placeholderText}>
              {unavailableReason ?? (permissionGranted ? "Camera ready" : "Camera permission required")}
            </Text>
          </View>
        )}
      </View>
      <Pressable
        style={[styles.captureButton, !permissionGranted ? styles.captureButtonDisabled : null]}
        onPress={handleCapture}
        disabled={!permissionGranted}
      >
        <Text style={styles.captureText}>Capture</Text>
      </Pressable>
    </View>
  );
}

function getCameraFailureReason(error: unknown): string {
  return error instanceof CameraUnavailableError ? error.code : "camera_unavailable";
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
  captureButtonDisabled: {
    opacity: 0.45
  },
  captureText: {
    color: "#ffffff",
    fontWeight: "600"
  }
});
