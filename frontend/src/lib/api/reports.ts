import { apiPost } from "./client";
import type { ReportData, ReportInput } from "./types";

export async function submitReport(input: ReportInput) {
  return apiPost<ReportData>("/api/v1/reports", input);
}
