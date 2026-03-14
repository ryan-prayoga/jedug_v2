import { derived, writable } from "svelte/store";
import {
  getNotifications,
  markNotificationRead,
  type Notification,
} from "$lib/api/notifications";
import { getOrCreateIssueFollowerId } from "$lib/utils/storage";

interface NotificationState {
  items: Notification[];
  loading: boolean;
  error: string | null;
  followerID: string | null;
  initialized: boolean;
}

const initialState: NotificationState = {
  items: [],
  loading: false,
  error: null,
  followerID: null,
  initialized: false,
};

const state = writable<NotificationState>(initialState);

export const notificationsState = {
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
    await this.refresh();
  },

  async refresh() {
    const followerID = getOrCreateIssueFollowerId();
    if (!followerID) {
      state.update((prev) => ({
        ...prev,
        loading: false,
        error: null,
        followerID: null,
      }));
      return;
    }

    state.update((prev) => ({
      ...prev,
      loading: true,
      error: null,
      followerID,
    }));
    try {
      const result = await getNotifications(followerID, 50);
      const items = result.data?.items ?? [];
      state.update((prev) => ({
        ...prev,
        items,
        loading: false,
        error: null,
        followerID,
      }));
    } catch {
      state.update((prev) => ({
        ...prev,
        loading: false,
        error: "Belum bisa memuat notifikasi.",
      }));
    }
  },

  async markRead(notificationID: string) {
    let followerID: string | null = null;
    state.update((prev) => {
      followerID = prev.followerID;
      return prev;
    });
    if (!followerID) {
      return;
    }

    try {
      await markNotificationRead(notificationID, followerID);
      state.update((prev) => ({
        ...prev,
        items: prev.items.map((item) =>
          item.id === notificationID
            ? { ...item, read_at: item.read_at ?? new Date().toISOString() }
            : item,
        ),
      }));
    } catch {
      // keep UX non-blocking for now
    }
  },
};

export const unreadNotificationCount = derived(
  notificationsState,
  ($state) => $state.items.filter((item) => !item.read_at).length,
);
