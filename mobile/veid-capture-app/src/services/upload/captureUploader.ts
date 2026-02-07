import type { CapturePayload } from "../../core/captureModels";

export interface UploadResult {
  success: boolean;
  error?: string;
}

export async function uploadCapture(payload: CapturePayload): Promise<UploadResult> {
  if (!payload.transport.uploadUrl) {
    return { success: false, error: "missing_upload_url" };
  }

  return { success: true };
}
