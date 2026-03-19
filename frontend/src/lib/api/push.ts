import { apiGet, apiPost } from "./client";
import type { ApiResponse } from "./types";
import { getAnonToken } from "$lib/utils/storage";

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

function withFollowerTokenHeader(followerToken: string) {
  return {
    deviceToken: getAnonToken() ?? undefined,
    headers: {
      "X-Follower-Token": followerToken,
    },
  };
}

export async function getBrowserPushStatus(
  followerToken: string,
): Promise<ApiResponse<BrowserPushStatusResponse>> {
  return apiGet<BrowserPushStatusResponse>(
    "/api/v1/push/status",
    withFollowerTokenHeader(followerToken),
  );
}

export async function subscribeBrowserPush(
  followerToken: string,
  subscription: BrowserPushSubscriptionPayload,
): Promise<ApiResponse<BrowserPushStatusResponse>> {
  return apiPost<BrowserPushStatusResponse>(
    "/api/v1/push/subscribe",
    { subscription },
    withFollowerTokenHeader(followerToken),
  );
}

export async function unsubscribeBrowserPush(
  followerToken: string,
  endpoint: string,
): Promise<ApiResponse<BrowserPushUnsubscribeResponse>> {
  return apiPost<BrowserPushUnsubscribeResponse>(
    "/api/v1/push/unsubscribe",
    { endpoint },
    withFollowerTokenHeader(followerToken),
  );
}
