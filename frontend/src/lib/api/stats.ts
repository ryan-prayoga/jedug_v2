import { apiGet } from "./client";
import type { PublicStats } from "./types";

export async function getPublicStats() {
  return apiGet<PublicStats>("/api/v1/stats");
}
