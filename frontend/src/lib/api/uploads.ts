import { apiPost, apiUploadBinary } from "./client";
import type { PresignData } from "./types";

export async function presignUpload(
  filename: string,
  contentType: string,
  sizeBytes: number,
) {
  return apiPost<PresignData>("/api/v1/uploads/presign", {
    filename,
    content_type: contentType,
    size_bytes: sizeBytes,
  });
}

export async function uploadFile(
  uploadUrl: string,
  file: Blob,
  contentType: string,
) {
  return apiUploadBinary(uploadUrl, file, contentType);
}
