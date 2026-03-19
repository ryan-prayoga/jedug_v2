import { apiPost, apiUploadBinary } from "./client";
import type { PresignData } from "./types";

export async function presignUpload(
  anonToken: string,
  filename: string,
  contentType: string,
  sizeBytes: number,
) {
  return apiPost<PresignData>("/api/v1/uploads/presign", {
    anon_token: anonToken,
    filename,
    content_type: contentType,
    size_bytes: sizeBytes,
  });
}

export async function uploadFile(
  uploadUrl: string,
  file: Blob,
  contentType: string,
  uploadMethod = "POST",
  headers: Record<string, string> = {},
) {
  return apiUploadBinary(uploadUrl, file, contentType, uploadMethod, headers);
}
