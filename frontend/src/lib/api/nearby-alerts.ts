import { apiDelete, apiGet, apiPatch, apiPost } from "./client";
import type { ApiResponse } from "./types";

export interface NearbyAlertSubscription {
  id: string;
  follower_id: string;
  label: string | null;
  latitude: number;
  longitude: number;
  radius_m: number;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface NearbyAlertCreateInput {
  label?: string;
  latitude: number;
  longitude: number;
  radius_m: number;
  enabled?: boolean;
}

export interface NearbyAlertPatch {
  label?: string;
  latitude?: number;
  longitude?: number;
  radius_m?: number;
  enabled?: boolean;
}

export async function getNearbyAlerts(
  followerToken: string,
): Promise<ApiResponse<NearbyAlertSubscription[]>> {
  const params = new URLSearchParams({ follower_token: followerToken });
  return apiGet<NearbyAlertSubscription[]>(`/api/v1/nearby-alerts?${params}`);
}

export async function createNearbyAlert(
  followerToken: string,
  input: NearbyAlertCreateInput,
): Promise<ApiResponse<NearbyAlertSubscription>> {
  return apiPost<NearbyAlertSubscription>("/api/v1/nearby-alerts", {
    follower_token: followerToken,
    ...input,
  });
}

export async function patchNearbyAlert(
  followerToken: string,
  id: string,
  patch: NearbyAlertPatch,
): Promise<ApiResponse<NearbyAlertSubscription>> {
  return apiPatch<NearbyAlertSubscription>(`/api/v1/nearby-alerts/${encodeURIComponent(id)}`, {
    follower_token: followerToken,
    ...patch,
  });
}

export async function deleteNearbyAlert(
  followerToken: string,
  id: string,
): Promise<ApiResponse<{ deleted: boolean }>> {
  return apiDelete<{ deleted: boolean }>(`/api/v1/nearby-alerts/${encodeURIComponent(id)}`, {
    follower_token: followerToken,
  });
}