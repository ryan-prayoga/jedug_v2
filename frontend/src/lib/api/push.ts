import { apiGet, apiPost } from "./client";
import type { ApiResponse } from "./types";

export interface BrowserPushStatusResponse {
  enabled: boolean;
  subscribed: boolean;
  subscription_count: number;
  vapid_public_key?: string;
}

export interface BrowserPushSubscriptionPayload {
  endpoint: string;
  keys: {
    p256dh: string;
    auth: string;
  };
}

export interface BrowserPushUnsubscribeResponse {
  enabled: boolean;
  subscribed: boolean;
  subscription_count: number;
  unsubscribed: boolean;
}

export async function getBrowserPushStatus(
  followerToken: string,
): Promise<ApiResponse<BrowserPushStatusResponse>> {
  const params = new URLSearchParams({ follower_token: followerToken });
  return apiGet<BrowserPushStatusResponse>(`/api/v1/push/status?${params}`);
}

export async function subscribeBrowserPush(
  followerToken: string,
  subscription: BrowserPushSubscriptionPayload,
): Promise<ApiResponse<BrowserPushStatusResponse>> {
  return apiPost<BrowserPushStatusResponse>("/api/v1/push/subscribe", {
    follower_token: followerToken,
    subscription,
  });
}

export async function unsubscribeBrowserPush(
  followerToken: string,
  endpoint: string,
): Promise<ApiResponse<BrowserPushUnsubscribeResponse>> {
  return apiPost<BrowserPushUnsubscribeResponse>("/api/v1/push/unsubscribe", {
    follower_token: followerToken,
    endpoint,
  });
}
