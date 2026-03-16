import { apiGet, apiPatch } from "./client";
import type { ApiResponse } from "./types";

export interface NotificationPreferences {
  follower_id: string;
  notifications_enabled: boolean;
  in_app_enabled: boolean;
  push_enabled: boolean;
  notify_on_photo_added: boolean;
  notify_on_status_updated: boolean;
  notify_on_severity_changed: boolean;
  notify_on_casualty_reported: boolean;
  notify_on_nearby_issue_created: boolean;
  created_at: string;
  updated_at: string;
}

export interface NotificationPreferencesPatch {
  notifications_enabled?: boolean;
  in_app_enabled?: boolean;
  push_enabled?: boolean;
  notify_on_photo_added?: boolean;
  notify_on_status_updated?: boolean;
  notify_on_severity_changed?: boolean;
  notify_on_casualty_reported?: boolean;
  notify_on_nearby_issue_created?: boolean;
}

export async function getNotificationPreferences(
  followerToken: string,
): Promise<ApiResponse<NotificationPreferences>> {
  const params = new URLSearchParams({ follower_token: followerToken });
  return apiGet<NotificationPreferences>(
    `/api/v1/notification-preferences?${params}`,
  );
}

export async function patchNotificationPreferences(
  followerToken: string,
  patch: NotificationPreferencesPatch,
): Promise<ApiResponse<NotificationPreferences>> {
  return apiPatch<NotificationPreferences>("/api/v1/notification-preferences", {
    follower_token: followerToken,
    ...patch,
  });
}
