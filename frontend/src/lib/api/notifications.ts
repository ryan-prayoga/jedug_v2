import { apiDelete, apiGet, apiPatch } from "./client";
import type { ApiResponse } from "./types";

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

export async function getNotifications(
  followerID: string,
  limit = 50,
): Promise<ApiResponse<NotificationList>> {
  const params = new URLSearchParams({
    follower_id: followerID,
    limit: String(limit),
  });
  return apiGet<NotificationList>(`/api/v1/notifications?${params}`);
}

export async function markNotificationRead(
  notificationID: string,
  followerID: string,
): Promise<ApiResponse<NotificationReadResult>> {
  const params = new URLSearchParams({ follower_id: followerID });
  return apiPatch<NotificationReadResult>(
    `/api/v1/notifications/${notificationID}/read?${params}`,
  );
}

export async function deleteNotification(
  notificationID: string,
  followerID: string,
): Promise<ApiResponse<NotificationDeleteResult>> {
  const params = new URLSearchParams({ follower_id: followerID });
  return apiDelete<NotificationDeleteResult>(
    `/api/v1/notifications/${notificationID}?${params}`,
  );
}
