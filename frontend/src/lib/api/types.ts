// API response types matching backend contracts

export interface ApiResponse<T = undefined> {
  success: boolean;
  message?: string;
  data?: T;
}

// Device
export interface BootstrapData {
  device_id: string;
  anon_token: string;
  is_new: boolean;
}

// Upload
export interface PresignData {
  object_key: string;
  upload_mode: string;
  upload_url: string;
  upload_method?: string;
  public_url: string;
  headers?: Record<string, string>;
}

// Report
export interface ReportData {
  issue_id: string;
  submission_id: string;
  is_new_issue: boolean;
}

export interface ReportMediaInput {
  object_key: string;
  mime_type: string;
  size_bytes: number;
  width: number | null;
  height: number | null;
  sha256: string | null;
  is_primary: boolean;
}

export interface ReportInput {
  anon_token: string;
  client_request_id?: string;
  latitude: number;
  longitude: number;
  gps_accuracy_m?: number;
  severity: number;
  note?: string;
  has_casualty: boolean;
  casualty_count: number;
  captured_at?: string;
  media: ReportMediaInput[];
}

// Issue
export interface Issue {
  id: string;
  status: string;
  verification_status: string;
  severity_current: number;
  severity_max: number;
  longitude: number;
  latitude: number;
  region_id: number | null;
  region_name: string | null;
  road_name: string | null;
  road_type: string | null;
  submission_count: number;
  photo_count: number;
  casualty_count: number;
  reaction_count: number;
  flag_count: number;
  first_seen_at: string;
  last_seen_at: string;
  created_at: string;
  updated_at: string;
}

export interface MediaItem {
  id: string;
  object_key: string;
  public_url: string;
  mime_type: string;
  size_bytes: number;
  width: number | null;
  height: number | null;
  blurhash: string | null;
  is_primary: boolean;
  created_at: string;
}

export interface SubmissionSummary {
  id: string;
  status: string;
  severity: number;
  has_casualty: boolean;
  note: string | null;
  reported_at: string;
}

export interface IssueDetail extends Issue {
  media: MediaItem[];
  recent_submissions: SubmissionSummary[];
}

// Admin types
export interface AdminLoginResponse {
  token: string;
}

export interface AdminMeResponse {
  username: string;
}

export interface AdminIssue {
  id: string;
  status: string;
  verification_status: string;
  severity_current: number;
  severity_max: number;
  longitude: number;
  latitude: number;
  region_id: number | null;
  region_name: string | null;
  road_name: string | null;
  road_type: string | null;
  submission_count: number;
  photo_count: number;
  casualty_count: number;
  reaction_count: number;
  flag_count: number;
  is_hidden: boolean;
  first_seen_at: string;
  last_seen_at: string;
  created_at: string;
  updated_at: string;
}

export interface AdminSubmissionSummary {
  id: string;
  device_id: string;
  device_is_banned: boolean;
  status: string;
  severity: number;
  has_casualty: boolean;
  note: string | null;
  reported_at: string;
}

export interface ModerationAction {
  id: number;
  action_type: string;
  target_type: string;
  target_id: string;
  admin_username: string | null;
  note: string | null;
  created_at: string;
}

export interface AdminIssueDetail extends AdminIssue {
  media: MediaItem[];
  submissions: AdminSubmissionSummary[];
  moderation_log: ModerationAction[];
}
