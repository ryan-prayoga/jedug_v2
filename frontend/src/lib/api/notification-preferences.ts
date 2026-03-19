import { apiGet, apiPatch } from "./client";
import type { ApiResponse } from "./types";
import { getAnonToken } from "$lib/utils/storage";

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

function withFollowerTokenHeader(followerToken: string) {
  return {
    deviceToken: getAnonToken() ?? undefined,
    headers: {
      "X-Follower-Token": followerToken,
    },
  };
}

export async function getNotificationPreferences(
  followerToken: string,
): Promise<ApiResponse<NotificationPreferences>> {
  return apiGet<NotificationPreferences>(
    "/api/v1/notification-preferences",
    withFollowerTokenHeader(followerToken),
  );
}

export async function patchNotificationPreferences(
  followerToken: string,
  patch: NotificationPreferencesPatch,
): Promise<ApiResponse<NotificationPreferences>> {
  return apiPatch<NotificationPreferences>(
    "/api/v1/notification-preferences",
    patch,
    withFollowerTokenHeader(followerToken),
  );
}
