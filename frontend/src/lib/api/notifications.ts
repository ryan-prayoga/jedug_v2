import { apiDelete, apiGet, apiPatch } from "./client";
import type { ApiResponse } from "./types";
import { getAnonToken } from "$lib/utils/storage";

export interface Notification {
  id: string;
  issue_id: string;
  event_id: number;
  type: string;
  title: string;
  message: string;
  created_at: string;
  read_at?: string | null;
}

export interface NotificationList {
  items: Notification[];
}

export interface NotificationReadResult {
  read_at: string;
}

export interface NotificationDeleteResult {
  deleted: boolean;
}

function withFollowerTokenHeader(followerToken: string) {
  return {
    deviceToken: getAnonToken() ?? undefined,
    headers: {
      "X-Follower-Token": followerToken,
    },
  };
}

export async function getNotifications(
  followerToken: string,
  limit = 50,
): Promise<ApiResponse<NotificationList>> {
  const params = new URLSearchParams({
    limit: String(limit),
  });
  return apiGet<NotificationList>(
    `/api/v1/notifications?${params}`,
    withFollowerTokenHeader(followerToken),
  );
}

export async function markNotificationRead(
  notificationID: string,
  followerToken: string,
): Promise<ApiResponse<NotificationReadResult>> {
  return apiPatch<NotificationReadResult>(
    `/api/v1/notifications/${notificationID}/read`,
    undefined,
    withFollowerTokenHeader(followerToken),
  );
}

export async function deleteNotification(
  notificationID: string,
  followerToken: string,
): Promise<ApiResponse<NotificationDeleteResult>> {
  return apiDelete<NotificationDeleteResult>(
    `/api/v1/notifications/${notificationID}`,
    undefined,
    withFollowerTokenHeader(followerToken),
  );
}
