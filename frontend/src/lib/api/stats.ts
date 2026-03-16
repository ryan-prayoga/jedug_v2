import { apiGet } from "./client";
import type { PublicStats, PublicStatsRegionOptionsData } from "./types";

interface GetPublicStatsOptions {
  provinceID?: number | null;
  regencyID?: number | null;
}

export async function getPublicStats(options: GetPublicStatsOptions = {}) {
  const query = new URLSearchParams();
  if (options.provinceID) {
    query.set("province_id", String(options.provinceID));
  }
  if (options.regencyID) {
    query.set("regency_id", String(options.regencyID));
  }

  const suffix = query.size > 0 ? `?${query.toString()}` : "";
  return apiGet<PublicStats>(`/api/v1/stats${suffix}`);
}

export async function getPublicStatsRegionOptions() {
  return apiGet<PublicStatsRegionOptionsData>("/api/v1/stats/regions/options");
}
