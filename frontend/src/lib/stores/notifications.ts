import { derived, writable } from "svelte/store";
import { PUBLIC_API_BASE_URL } from "$env/static/public";
import {
  deleteNotification,
  getNotifications,
  markNotificationRead,
  type Notification,
} from "$lib/api/notifications";
import { ensureFollowerAuthToken } from "$lib/utils/follower-auth";
import { getOrCreateIssueFollowerId } from "$lib/utils/storage";

interface NotificationState {
  items: Notification[];
  loading: boolean;
  error: string | null;
  followerID: string | null;
  followerToken: string | null;
  initialized: boolean;
  deletingIDs: string[];
}

type CrossTabMessage =
  | { type: "mark-read"; notificationID: string; readAt: string }
  | { type: "delete"; notificationID: string };

const initialState: NotificationState = {
  items: [],
  loading: false,
  error: null,
  followerID: null,
  followerToken: null,
  initialized: false,
  deletingIDs: [],
};

const state = writable<NotificationState>(initialState);

let _es: EventSource | null = null;
let _reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let _reconnectDelay = 1_000;
const _maxReconnectDelay = 30_000;
let _reconnectAttempts = 0;
let _realtimeEnabled = true;
const NOTIFICATION_SYNC_KEY = "jedug_notification_sync";
const channel =
  typeof window !== "undefined" && "BroadcastChannel" in window
    ? new BroadcastChannel("jedug-notifications")
    : null;

function applyMarkRead(notificationID: string, readAt: string) {
  state.update((prev) => ({
    ...prev,
    items: prev.items.map((item) =>
      item.id === notificationID
        ? { ...item, read_at: item.read_at ?? readAt }
        : item,
    ),
  }));
}

function applyDelete(notificationID: string) {
  state.update((prev) => ({
    ...prev,
    deletingIDs: prev.deletingIDs.filter((id) => id !== notificationID),
    items: prev.items.filter((item) => item.id !== notificationID),
  }));
}

function publishCrossTab(message: CrossTabMessage) {
  if (typeof window === "undefined") return;

  if (channel) {
    channel.postMessage(message);
    return;
  }

  localStorage.setItem(
    NOTIFICATION_SYNC_KEY,
    JSON.stringify({ ...message, ts: Date.now() }),
  );
  localStorage.removeItem(NOTIFICATION_SYNC_KEY);
}

function handleCrossTabMessage(message: CrossTabMessage) {
  if (message.type === "mark-read") {
    applyMarkRead(message.notificationID, message.readAt);
    return;
  }

  applyDelete(message.notificationID);
}

if (typeof window !== "undefined") {
  channel?.addEventListener("message", (event) => {
    const data = event.data as CrossTabMessage | undefined;
    if (!data || typeof data !== "object") return;
    handleCrossTabMessage(data);
  });

  window.addEventListener("storage", (event) => {
    if (event.key !== NOTIFICATION_SYNC_KEY || !event.newValue) return;
    try {
      const parsed = JSON.parse(event.newValue) as CrossTabMessage;
      handleCrossTabMessage(parsed);
    } catch {
      // ignore malformed sync payload
    }
  });
}

function clearReconnectTimer() {
  if (_reconnectTimer !== null) {
    clearTimeout(_reconnectTimer);
    _reconnectTimer = null;
  }
}

function setSnapshot(
  followerID: string,
  followerToken: string,
  items: Notification[],
  loading: boolean,
) {
  state.update((prev) => ({
    ...prev,
    items,
    loading,
    error: null,
    followerID,
    followerToken,
  }));
}

async function fetchSnapshot(
  followerID: string,
  followerToken: string,
  loading: boolean,
) {
  if (loading) {
    state.update((prev) => ({
      ...prev,
      loading: true,
      error: null,
      followerID,
      followerToken,
    }));
  }

  const result = await getNotifications(followerToken, 50);
  setSnapshot(followerID, followerToken, result.data?.items ?? [], false);
}

function scheduleReconnect(followerID: string) {
  clearReconnectTimer();

  const delay = _reconnectDelay;
  _reconnectDelay = Math.min(_reconnectDelay * 2, _maxReconnectDelay);

  _reconnectTimer = setTimeout(async () => {
    _reconnectTimer = null;

    try {
      if (_reconnectAttempts >= 3) {
        await refreshState(false, false, true);
      }
    } catch {
      // best-effort fallback refresh
    }

    const latestToken = await ensureFollowerAuthToken();
    if (!latestToken) {
      _disconnectSSE();
      return;
    }

    _connectSSE(followerID, latestToken);
  }, delay);
}

function _connectSSE(followerID: string, followerToken: string) {
  if (typeof EventSource === "undefined") return;

  clearReconnectTimer();
  if (_es) {
    _es.close();
    _es = null;
  }

  const params = new URLSearchParams({ follower_token: followerToken });
  const es = new EventSource(
    `${PUBLIC_API_BASE_URL}/api/v1/notifications/stream?${params.toString()}`,
  );
  _es = es;

  es.addEventListener("connected", () => {
    _reconnectDelay = 1_000;
    _reconnectAttempts = 0;
  });

  es.addEventListener("notification", (e: MessageEvent) => {
    try {
      const incoming = JSON.parse(e.data) as Notification;
      state.update((prev) => {
        if (prev.items.some((item) => item.id === incoming.id)) return prev;
        return { ...prev, items: [incoming, ...prev.items] };
      });
      _reconnectDelay = 1_000;
      _reconnectAttempts = 0;
    } catch {
      // ignore malformed notification payload
    }
  });

  es.addEventListener("ping", () => {});

  es.onerror = () => {
    es.close();
    if (_es === es) {
      _es = null;
    }

    _reconnectAttempts += 1;
    scheduleReconnect(followerID);
  };
}

function _disconnectSSE() {
  clearReconnectTimer();
  if (_es) {
    _es.close();
    _es = null;
  }
  _reconnectDelay = 1_000;
  _reconnectAttempts = 0;
}

async function refreshState(
  showLoading: boolean,
  reconnectStream = true,
  forceAuthRefresh = false,
) {
  const followerID = getOrCreateIssueFollowerId();
  if (!followerID) {
    _disconnectSSE();
    state.update((prev) => ({
      ...prev,
      loading: false,
      error: null,
      followerID: null,
      followerToken: null,
      items: [],
    }));
    return;
  }

  const followerToken = await ensureFollowerAuthToken({
    forceRefresh: forceAuthRefresh,
  });

  if (!followerToken) {
    _disconnectSSE();
    state.update((prev) => ({
      ...prev,
      loading: false,
      error: null,
      followerID,
      followerToken: null,
      items: [],
    }));
    return;
  }

  try {
    await fetchSnapshot(followerID, followerToken, showLoading);
    if (reconnectStream && _realtimeEnabled) {
      _connectSSE(followerID, followerToken);
    } else if (!_realtimeEnabled) {
      _disconnectSSE();
    }
  } catch {
    state.update((prev) => ({
      ...prev,
      loading: false,
      error: "Belum bisa memuat notifikasi.",
      followerID,
      followerToken,
    }));
  }
}

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
    await refreshState(true, true);
  },

  async refresh() {
    await refreshState(false, true);
  },

  async markRead(notificationID: string) {
    let followerToken: string | null = null;
    state.update((prev) => {
      followerToken = prev.followerToken;
      return prev;
    });

    followerToken =
      followerToken ?? (await ensureFollowerAuthToken().catch(() => null));
    if (!followerToken) {
      return;
    }

    try {
      const result = await markNotificationRead(notificationID, followerToken);
      const readAt = result.data?.read_at ?? new Date().toISOString();
      applyMarkRead(notificationID, readAt);
      publishCrossTab({ type: "mark-read", notificationID, readAt });
    } catch {
      state.update((prev) => ({
        ...prev,
        error: "Gagal sinkronisasi status dibaca. Coba lagi.",
      }));
      await refreshState(false, false, true);
    }
  },

  async delete(notificationID: string) {
    let followerToken: string | null = null;
    let shouldDelete = false;
    state.update((prev) => {
      followerToken = prev.followerToken;
      if (prev.deletingIDs.includes(notificationID)) {
        return prev;
      }

      shouldDelete = true;
      return {
        ...prev,
        error: null,
        deletingIDs: [...prev.deletingIDs, notificationID],
      };
    });

    if (!shouldDelete) {
      return false;
    }

    followerToken =
      followerToken ?? (await ensureFollowerAuthToken().catch(() => null));
    if (!followerToken) {
      state.update((prev) => ({
        ...prev,
        deletingIDs: prev.deletingIDs.filter((id) => id !== notificationID),
      }));
      return false;
    }

    try {
      const result = await deleteNotification(notificationID, followerToken);
      if (result.data?.deleted) {
        applyDelete(notificationID);
        publishCrossTab({ type: "delete", notificationID });
      } else {
        state.update((prev) => ({
          ...prev,
          deletingIDs: prev.deletingIDs.filter((id) => id !== notificationID),
        }));
      }

      return Boolean(result.data?.deleted);
    } catch {
      state.update((prev) => ({
        ...prev,
        error: "Belum bisa menghapus notifikasi. Coba lagi.",
        deletingIDs: prev.deletingIDs.filter((id) => id !== notificationID),
      }));
      await refreshState(false, false, true);
      return false;
    }
  },

  disconnect: _disconnectSSE,

  setRealtimeEnabled(enabled: boolean) {
    _realtimeEnabled = enabled;
    if (!enabled) {
      _disconnectSSE();
      return;
    }

    let shouldRefresh = false;
    state.update((prev) => {
      shouldRefresh =
        prev.initialized &&
        Boolean(prev.followerID) &&
        Boolean(prev.followerToken);
      return prev;
    });

    if (shouldRefresh) {
      void refreshState(false, true);
    }
  },
};

export const unreadNotificationCount = derived(
  notificationsState,
  ($state) => $state.items.filter((item) => !item.read_at).length,
);
