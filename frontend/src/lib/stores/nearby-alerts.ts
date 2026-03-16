import { writable } from "svelte/store";
import { ApiError } from "$lib/api/client";
import {
  createNearbyAlert,
  deleteNearbyAlert,
  getNearbyAlerts,
  patchNearbyAlert,
  type NearbyAlertCreateInput,
  type NearbyAlertPatch,
  type NearbyAlertSubscription,
} from "$lib/api/nearby-alerts";
import { ensureFollowerAuthToken } from "$lib/utils/follower-auth";
import { getOrCreateIssueFollowerId } from "$lib/utils/storage";

interface NearbyAlertsState {
  items: NearbyAlertSubscription[];
  loading: boolean;
  creating: boolean;
  savingKeys: string[];
  deletingIDs: string[];
  initialized: boolean;
  followerID: string | null;
  followerToken: string | null;
  error: string | null;
  unavailableMessage: string | null;
}

const initialState: NearbyAlertsState = {
  items: [],
  loading: false,
  creating: false,
  savingKeys: [],
  deletingIDs: [],
  initialized: false,
  followerID: null,
  followerToken: null,
  error: null,
  unavailableMessage: null,
};

const state = writable<NearbyAlertsState>(initialState);

function humanizeNearbyAlertError(error: unknown, fallback: string): string {
  if (error instanceof ApiError) {
    if (error.errorCode === "nearby_alert_limit_exceeded") {
      return "Maksimum 10 lokasi pantauan per browser agar notifikasi tetap ringan.";
    }

    switch (error.message) {
      case "nearby alert coordinates are invalid":
        return "Koordinat belum valid. Periksa latitude dan longitude lagi.";
      case "latitude and longitude must be provided together":
        return "Latitude dan longitude harus diisi bersamaan.";
      case "nearby alert radius is invalid":
        return "Radius harus antara 100 sampai 5000 meter.";
      case "nearby alert label is too long":
        return "Label lokasi terlalu panjang. Maksimum 80 karakter.";
      case "nearby alert subscription not found":
        return "Lokasi pantauan ini sudah tidak ditemukan.";
      case "at least one nearby alert field must be provided":
        return "Belum ada perubahan yang bisa disimpan.";
      default:
        return error.message || fallback;
    }
  }

  if (error instanceof Error && error.message) {
    return error.message;
  }

  return fallback;
}

async function resolveFollowerContext(forceAuthRefresh = false) {
  const followerID = getOrCreateIssueFollowerId();
  if (!followerID) {
    state.update((prev) => ({
      ...prev,
      loading: false,
      followerID: null,
      followerToken: null,
      unavailableMessage: "Identitas browser belum siap.",
    }));
    return null;
  }

  const followerToken = await ensureFollowerAuthToken({
    forceRefresh: forceAuthRefresh,
  });
  if (!followerToken) {
    state.update((prev) => ({
      ...prev,
      loading: false,
      followerID,
      followerToken: null,
      unavailableMessage:
        "Sesi notifikasi belum siap. Coba lagi sebentar lagi.",
    }));
    return null;
  }

  return { followerID, followerToken };
}

async function refreshState(showLoading: boolean, forceAuthRefresh = false) {
  if (showLoading) {
    state.update((prev) => ({ ...prev, loading: true, error: null }));
  }

  const context = await resolveFollowerContext(forceAuthRefresh);
  if (!context) return;

  try {
    const result = await getNearbyAlerts(context.followerToken);
    state.update((prev) => ({
      ...prev,
      items: result.data ?? [],
      loading: false,
      error: null,
      unavailableMessage: null,
      followerID: context.followerID,
      followerToken: context.followerToken,
    }));
  } catch (error) {
    state.update((prev) => ({
      ...prev,
      loading: false,
      error: humanizeNearbyAlertError(
        error,
        "Belum bisa memuat lokasi pantauan.",
      ),
      followerID: context.followerID,
      followerToken: context.followerToken,
    }));
  }
}

export const nearbyAlertsState = {
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

  async create(input: NearbyAlertCreateInput): Promise<boolean> {
    let followerToken: string | null = null;
    state.update((prev) => {
      followerToken = prev.followerToken;
      return { ...prev, creating: true, error: null };
    });

    followerToken =
      followerToken ??
      (await ensureFollowerAuthToken({ forceRefresh: true }).catch(() => null));
    if (!followerToken) {
      state.update((prev) => ({
        ...prev,
        creating: false,
        unavailableMessage: "Sesi notifikasi perlu diperbarui. Coba lagi.",
      }));
      return false;
    }

    try {
      const result = await createNearbyAlert(followerToken, input);
      const created = result.data;
      if (!created) {
        throw new Error("missing nearby alert subscription");
      }
      state.update((prev) => ({
        ...prev,
        creating: false,
        followerToken,
        items: [
          created,
          ...prev.items.filter((item) => item.id !== created.id),
        ],
        error: null,
      }));
      return true;
    } catch (error) {
      state.update((prev) => ({
        ...prev,
        creating: false,
        error: humanizeNearbyAlertError(
          error,
          "Belum bisa menambah lokasi pantauan.",
        ),
      }));
      return false;
    }
  },

  async update(
    id: string,
    patch: NearbyAlertPatch,
    savingKey?: string,
  ): Promise<boolean> {
    let followerToken: string | null = null;
    let previousItems: NearbyAlertSubscription[] = [];
    let shouldContinue = true;

    state.update((prev) => {
      if (savingKey && prev.savingKeys.includes(savingKey)) {
        shouldContinue = false;
        return prev;
      }
      followerToken = prev.followerToken;
      previousItems = prev.items;

      return {
        ...prev,
        error: null,
        savingKeys: savingKey
          ? [...prev.savingKeys, savingKey]
          : prev.savingKeys,
        items: prev.items.map((item) =>
          item.id === id ? { ...item, ...patch } : item,
        ),
      };
    });

    if (!shouldContinue) return false;

    followerToken =
      followerToken ??
      (await ensureFollowerAuthToken({ forceRefresh: true }).catch(() => null));
    if (!followerToken) {
      state.update((prev) => ({
        ...prev,
        items: previousItems,
        savingKeys: savingKey
          ? prev.savingKeys.filter((key) => key !== savingKey)
          : prev.savingKeys,
        unavailableMessage: "Sesi notifikasi perlu diperbarui. Coba lagi.",
      }));
      return false;
    }

    try {
      const result = await patchNearbyAlert(followerToken, id, patch);
      const updated = result.data;
      if (!updated) {
        throw new Error("missing nearby alert subscription");
      }
      state.update((prev) => ({
        ...prev,
        followerToken,
        items: prev.items.map((item) => (item.id === id ? updated : item)),
        savingKeys: savingKey
          ? prev.savingKeys.filter((key) => key !== savingKey)
          : prev.savingKeys,
      }));
      return true;
    } catch (error) {
      state.update((prev) => ({
        ...prev,
        items: previousItems,
        error: humanizeNearbyAlertError(
          error,
          "Belum bisa menyimpan perubahan lokasi pantauan.",
        ),
        savingKeys: savingKey
          ? prev.savingKeys.filter((key) => key !== savingKey)
          : prev.savingKeys,
      }));
      return false;
    }
  },

  async remove(id: string): Promise<boolean> {
    let followerToken: string | null = null;
    let previousItems: NearbyAlertSubscription[] = [];

    state.update((prev) => {
      followerToken = prev.followerToken;
      previousItems = prev.items;
      return {
        ...prev,
        error: null,
        deletingIDs: [...prev.deletingIDs, id],
        items: prev.items.filter((item) => item.id !== id),
      };
    });

    followerToken =
      followerToken ??
      (await ensureFollowerAuthToken({ forceRefresh: true }).catch(() => null));
    if (!followerToken) {
      state.update((prev) => ({
        ...prev,
        items: previousItems,
        deletingIDs: prev.deletingIDs.filter((itemID) => itemID !== id),
        unavailableMessage: "Sesi notifikasi perlu diperbarui. Coba lagi.",
      }));
      return false;
    }

    try {
      const result = await deleteNearbyAlert(followerToken, id);
      const deleted = Boolean(result.data?.deleted);
      state.update((prev) => ({
        ...prev,
        followerToken,
        deletingIDs: prev.deletingIDs.filter((itemID) => itemID !== id),
        items: deleted ? prev.items : previousItems,
      }));
      return deleted;
    } catch (error) {
      state.update((prev) => ({
        ...prev,
        items: previousItems,
        error: humanizeNearbyAlertError(
          error,
          "Belum bisa menghapus lokasi pantauan.",
        ),
        deletingIDs: prev.deletingIDs.filter((itemID) => itemID !== id),
      }));
      return false;
    }
  },
};
