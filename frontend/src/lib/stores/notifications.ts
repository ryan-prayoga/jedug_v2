import { derived, writable } from "svelte/store";
import { PUBLIC_API_BASE_URL } from "$env/static/public";
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

// ── SSE state (module-level, not part of the Svelte store) ───────────────────
let _es: EventSource | null = null;
let _reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let _reconnectDelay = 1_000; // ms, doubles on each failure up to MAX
const _maxReconnectDelay = 30_000;
let _reconnectAttempts = 0;
const _maxReconnectAttempts = 10; // give up after 10 consecutive failures

function _connectSSE(followerID: string) {
  if (typeof EventSource === "undefined") return; // SSR or unsupported browser

  if (_es) {
    _es.close();
    _es = null;
  }

  const url = `${PUBLIC_API_BASE_URL}/api/v1/notifications/stream?follower_id=${encodeURIComponent(followerID)}`;
  const es = new EventSource(url);
  _es = es;

  es.addEventListener("connected", () => {
    // Reset backoff on successful connection
    _reconnectDelay = 1_000;
    _reconnectAttempts = 0;
  });

  es.addEventListener("notification", (e: MessageEvent) => {
    try {
      const incoming = JSON.parse(e.data) as Notification;
      state.update((prev) => {
        // Deduplicate by id
        if (prev.items.some((item) => item.id === incoming.id)) return prev;
        return { ...prev, items: [incoming, ...prev.items] };
      });
      // Also reset backoff — we're clearly connected fine
      _reconnectDelay = 1_000;
      _reconnectAttempts = 0;
    } catch {
      // malformed event — ignore
    }
  });

  // ping events are heartbeats; no action needed
  es.addEventListener("ping", () => {});

  es.onerror = () => {
    es.close();
    _es = null;

    _reconnectAttempts++;
    if (_reconnectAttempts > _maxReconnectAttempts) {
      // Stop reconnecting — fall back to the already-fetched snapshot.
      // The user can manually refresh or the next page load will retry.
      return;
    }

    const delay = _reconnectDelay;
    _reconnectDelay = Math.min(_reconnectDelay * 2, _maxReconnectDelay);

    _reconnectTimer = setTimeout(() => {
      _reconnectTimer = null;
      _connectSSE(followerID);
    }, delay);
  };
}

function _disconnectSSE() {
  if (_reconnectTimer !== null) {
    clearTimeout(_reconnectTimer);
    _reconnectTimer = null;
  }
  if (_es) {
    _es.close();
    _es = null;
  }
  _reconnectDelay = 1_000;
  _reconnectAttempts = 0;
}

// ── Public store API ─────────────────────────────────────────────────────────

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

    // Open SSE stream after the initial fetch so we don't miss events that
    // arrive between page load and the HTTP response.
    _connectSSE(followerID);
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

  /** Disconnect the SSE stream (call on page/component destroy if needed). */
  disconnect: _disconnectSSE,
};

export const unreadNotificationCount = derived(
  notificationsState,
  ($state) => $state.items.filter((item) => !item.read_at).length,
);
