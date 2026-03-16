import { writable } from "svelte/store";
import {
  getNotificationPreferences,
  patchNotificationPreferences,
  type NotificationPreferences,
  type NotificationPreferencesPatch,
} from "$lib/api/notification-preferences";
import { notificationsState } from "$lib/stores/notifications";
import {
  ensureFollowerAuthToken,
  getFollowerAuthUnavailableMessage,
} from "$lib/utils/follower-auth";
import { getOrCreateIssueFollowerId } from "$lib/utils/storage";

interface NotificationPreferencesState {
  preferences: NotificationPreferences | null;
  loading: boolean;
  savingKeys: string[];
  error: string | null;
  initialized: boolean;
  followerID: string | null;
  followerToken: string | null;
  unavailableMessage: string | null;
}

type CrossTabMessage = {
  type: "snapshot";
  preferences: NotificationPreferences;
};

const initialState: NotificationPreferencesState = {
  preferences: null,
  loading: false,
  savingKeys: [],
  error: null,
  initialized: false,
  followerID: null,
  followerToken: null,
  unavailableMessage: null,
};

const state = writable<NotificationPreferencesState>(initialState);

const NOTIFICATION_PREFERENCES_SYNC_KEY = "jedug_notification_preferences_sync";
const channel =
  typeof window !== "undefined" && "BroadcastChannel" in window
    ? new BroadcastChannel("jedug-notification-preferences")
    : null;

function syncRealtimePreference(preferences: NotificationPreferences | null) {
  const enabled = Boolean(
    preferences?.notifications_enabled && preferences.in_app_enabled,
  );
  notificationsState.setRealtimeEnabled(enabled);
}

function applySnapshot(preferences: NotificationPreferences) {
  syncRealtimePreference(preferences);
  state.update((prev) => ({
    ...prev,
    preferences,
    loading: false,
    error: null,
    unavailableMessage: null,
  }));
}

function publishCrossTab(message: CrossTabMessage) {
  if (typeof window === "undefined") return;

  if (channel) {
    channel.postMessage(message);
    return;
  }

  localStorage.setItem(
    NOTIFICATION_PREFERENCES_SYNC_KEY,
    JSON.stringify({ ...message, ts: Date.now() }),
  );
  localStorage.removeItem(NOTIFICATION_PREFERENCES_SYNC_KEY);
}

function handleCrossTabMessage(message: CrossTabMessage) {
  if (message.type !== "snapshot") return;
  applySnapshot(message.preferences);
}

if (typeof window !== "undefined") {
  channel?.addEventListener("message", (event) => {
    const data = event.data as CrossTabMessage | undefined;
    if (!data || typeof data !== "object") return;
    handleCrossTabMessage(data);
  });

  window.addEventListener("storage", (event) => {
    if (
      event.key !== NOTIFICATION_PREFERENCES_SYNC_KEY ||
      !event.newValue
    ) {
      return;
    }
    try {
      const parsed = JSON.parse(event.newValue) as CrossTabMessage;
      handleCrossTabMessage(parsed);
    } catch {
      // ignore malformed sync payload
    }
  });
}

async function refreshState(showLoading: boolean, forceAuthRefresh = false) {
  const followerID = getOrCreateIssueFollowerId();
  if (!followerID) {
    state.update((prev) => ({
      ...prev,
      loading: false,
      preferences: null,
      followerID: null,
      followerToken: null,
      unavailableMessage: "Identitas browser belum siap.",
    }));
    return;
  }

  const followerToken = await ensureFollowerAuthToken({
    forceRefresh: forceAuthRefresh,
  });
  if (!followerToken) {
    state.update((prev) => ({
      ...prev,
      loading: false,
      error: null,
      preferences: null,
      followerID,
      followerToken: null,
      unavailableMessage: getFollowerAuthUnavailableMessage(
        "Ikuti setidaknya satu laporan di browser ini untuk mengatur notifikasi.",
      ),
    }));
    return;
  }

  if (showLoading) {
    state.update((prev) => ({
      ...prev,
      loading: true,
      error: null,
      unavailableMessage: null,
      followerID,
      followerToken,
    }));
  }

  try {
    const result = await getNotificationPreferences(followerToken);
    const preferences = result.data ?? null;
    if (!preferences) {
      throw new Error("missing notification preferences");
    }

    syncRealtimePreference(preferences);
    state.update((prev) => ({
      ...prev,
      preferences,
      loading: false,
      error: null,
      unavailableMessage: null,
      followerID,
      followerToken,
    }));
  } catch {
    state.update((prev) => ({
      ...prev,
      loading: false,
      error: "Belum bisa memuat preferensi notifikasi.",
      followerID,
      followerToken,
    }));
  }
}

function isSavingKeyInFlight(keys: string[], savingKey?: string) {
  return Boolean(savingKey && keys.includes(savingKey));
}

export const notificationPreferencesState = {
  subscribe: state.subscribe,

  async init() {
    let shouldInit = false;
    state.update((prev) => {
      if (!prev.initialized) {
        shouldInit = true;
        return { ...prev, initialized: true };
      }
      return prev;
    });

    if (!shouldInit) return;
    await refreshState(true, false);
  },

  async refresh(forceAuthRefresh = false) {
    await refreshState(false, forceAuthRefresh);
  },

  async update(
    patch: NotificationPreferencesPatch,
    savingKey?: string,
  ): Promise<boolean> {
    let followerToken: string | null = null;
    let currentPreferences: NotificationPreferences | null = null;
    let shouldContinue = true;

    state.update((prev) => {
      followerToken = prev.followerToken;
      currentPreferences = prev.preferences;

      if (!prev.preferences || isSavingKeyInFlight(prev.savingKeys, savingKey)) {
        shouldContinue = false;
        return prev;
      }

      return {
        ...prev,
        error: null,
        preferences: {
          ...prev.preferences,
          ...patch,
        },
        savingKeys: savingKey
          ? [...prev.savingKeys, savingKey]
          : prev.savingKeys,
      };
    });

    if (!shouldContinue) {
      return false;
    }

    followerToken =
      followerToken ??
      (await ensureFollowerAuthToken({ forceRefresh: true }).catch(() => null));
    if (!followerToken) {
      state.update((prev) => ({
        ...prev,
        preferences: currentPreferences,
        savingKeys: savingKey
          ? prev.savingKeys.filter((key) => key !== savingKey)
          : prev.savingKeys,
        unavailableMessage: getFollowerAuthUnavailableMessage(
          "Sesi notifikasi perlu diperbarui. Coba lagi dari browser ini.",
        ),
      }));
      return false;
    }

    try {
      const result = await patchNotificationPreferences(followerToken, patch);
      const preferences = result.data ?? null;
      if (!preferences) {
        throw new Error("missing notification preferences");
      }

      applySnapshot(preferences);
      publishCrossTab({ type: "snapshot", preferences });
      state.update((prev) => ({
        ...prev,
        followerToken,
        savingKeys: savingKey
          ? prev.savingKeys.filter((key) => key !== savingKey)
          : prev.savingKeys,
      }));
      return true;
    } catch {
      syncRealtimePreference(currentPreferences);
      state.update((prev) => ({
        ...prev,
        preferences: currentPreferences,
        error: "Belum bisa menyimpan preferensi notifikasi.",
        savingKeys: savingKey
          ? prev.savingKeys.filter((key) => key !== savingKey)
          : prev.savingKeys,
      }));
      return false;
    }
  },
};
