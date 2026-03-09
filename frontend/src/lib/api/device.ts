import { apiPost } from "./client";
import type { BootstrapData } from "./types";

export async function bootstrapDevice(token?: string) {
  return apiPost<BootstrapData>("/api/v1/device/bootstrap", undefined, token);
}

export async function recordConsent(
  anonToken: string,
  termsVersion = "1.0",
  privacyVersion = "1.0",
) {
  return apiPost<undefined>("/api/v1/device/consent", {
    anon_token: anonToken,
    terms_version: termsVersion,
    privacy_version: privacyVersion,
  });
}
