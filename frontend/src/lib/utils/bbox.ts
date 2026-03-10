import type { ListIssuesParams } from "$lib/api/issues";
import type { Issue } from "$lib/api/types";
import { listIssues } from "$lib/api/issues";

export type BBox = [number, number, number, number]; // [minLng, minLat, maxLng, maxLat]

let debounceTimer: ReturnType<typeof setTimeout> | null = null;
let lastBBoxKey = "";
let activeRequestId = 0;

/**
 * Fetch issues for a given bounding box with debounce.
 * Returns null if the request was debounced (a newer request will come).
 */
export function fetchIssuesByBBox(
  bbox: BBox,
  options: { limit?: number; status?: string; severity?: number } = {},
  callback: (issues: Issue[], error: string | null) => void,
  debounceMs = 300,
): "scheduled" | "skipped" {
  const key = bbox.map((v) => v.toFixed(5)).join(",");

  // Skip if same bbox
  if (key === lastBBoxKey) return "skipped";
  lastBBoxKey = key;

  if (debounceTimer) clearTimeout(debounceTimer);

  debounceTimer = setTimeout(async () => {
    const requestId = ++activeRequestId;
    try {
      const params: ListIssuesParams = {
        bbox,
        limit: options.limit ?? 100,
        ...(options.status ? { status: options.status } : {}),
        ...(options.severity ? { severity: options.severity } : {}),
      };
      const res = await listIssues(params);
      if (requestId === activeRequestId) {
        callback(res.data || [], null);
      }
    } catch (e) {
      if (requestId === activeRequestId) {
        callback([], e instanceof Error ? e.message : "Gagal memuat data peta");
      }
    }
  }, debounceMs);

  return "scheduled";
}

/** Reset debounce state (e.g. on component destroy) */
export function resetBBoxFetcher(): void {
  if (debounceTimer) clearTimeout(debounceTimer);
  lastBBoxKey = "";
  activeRequestId = 0;
}
