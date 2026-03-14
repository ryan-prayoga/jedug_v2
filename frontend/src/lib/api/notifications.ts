import { apiGet, apiPatch } from "./client";
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
): Promise<ApiResponse<undefined>> {
  const params = new URLSearchParams({ follower_id: followerID });
  return apiPatch<undefined>(
    `/api/v1/notifications/${notificationID}/read?${params}`,
  );
}
