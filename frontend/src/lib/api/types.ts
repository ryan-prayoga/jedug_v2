// API response types matching backend contracts

export interface ApiResponse<T = undefined> {
  success: boolean;
  error_code?: string;
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
  upload_token: string;
  upload_expires_at: string;
  upload_token_type: string;
  upload_token_header?: string;
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
  upload_token: string;
  width: number | null;
  height: number | null;
  sha256: string | null;
  is_primary: boolean;
}

export interface ReportInput {
  anon_token: string;
  actor_follower_id?: string;
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

// Location label
export interface LocationLabelData {
  label: string | null;
  region_id: number | null;
  region_name: string | null;
  region_level: string | null;
  parent_name: string | null;
  grandparent_name: string | null;
  district_name: string | null;
  regency_name: string | null;
  province_name: string | null;
  source: string;
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
  casualty_count: number;
  note: string | null;
  public_note?: string | null;
  reported_at: string;
}

export interface IssueDetail extends Issue {
  primary_media?: MediaItem | null;
  public_note?: string | null;
  media: MediaItem[];
  recent_submissions: SubmissionSummary[];
}

export interface IssueFollowState {
  following: boolean;
  followers_count: number;
  follower_token?: string;
  follower_token_expires_at?: string;
  follower_stream_token?: string;
  follower_stream_token_expires_at?: string;
}

export interface NearbyAlertSubscription {
  id: string;
  follower_id: string;
  label: string | null;
  latitude: number;
  longitude: number;
  radius_m: number;
  enabled: boolean;
  created_at: string;
  updated_at: string;
}

export interface IssueFollowersCount {
  followers_count: number;
}

export interface IssueTimelineEvent {
  type: string;
  created_at: string;
  data: Record<string, unknown>;
}

// Public stats
export interface PublicStatsGlobal {
  total_issues: number;
  total_issues_this_week: number;
  total_casualties: number;
  total_photos: number;
  total_reports: number;
}

export interface PublicStatsScope {
  kind: "global" | "province" | "regency";
  label: string;
  is_default: boolean;
}

export interface PublicStatsStatus {
  open: number;
  fixed: number;
  archived: number;
}

export interface PublicStatsTime {
  average_issue_age_days: number;
  oldest_open_issue_age_days: number;
  oldest_open_issue_id?: string | null;
  oldest_open_road_name?: string | null;
  oldest_open_region_name?: string | null;
  oldest_open_first_seen_at?: string | null;
}

export interface PublicStatsRegion {
  region_id: number;
  region_level: string;
  region_name: string;
  parent_region_name?: string | null;
  district_id?: number | null;
  district_name: string;
  regency_name?: string | null;
  province_name?: string | null;
  issue_count: number;
  casualty_count: number;
  report_count: number;
}

export interface PublicStatsRegionOption {
  id: number;
  name: string;
  issue_count: number;
  report_count: number;
}

export interface PublicStatsProvinceOption {
  id: number;
  name: string;
  issue_count: number;
  report_count: number;
  regencies: PublicStatsRegionOption[];
}

export interface PublicStatsRegionOptionsData {
  provinces: PublicStatsProvinceOption[];
}

export interface PublicStatsFilters {
  province_options: PublicStatsRegionOption[];
  regency_options: PublicStatsRegionOption[];
  active_province_id?: number | null;
  active_province?: string | null;
  active_regency_id?: number | null;
  active_regency?: string | null;
  scope_label?: string | null;
}

export interface PublicTopIssue {
  category: string;
  label: string;
  metric_label: string;
  metric_value: number;
  issue_id: string;
  status: string;
  road_name?: string | null;
  region_name?: string | null;
  district_name?: string | null;
  regency_name?: string | null;
  province_name?: string | null;
  submission_count: number;
  casualty_count: number;
  age_days: number;
}

export interface PublicStats {
  global: PublicStatsGlobal;
  summary: PublicStatsGlobal;
  status: PublicStatsStatus;
  time: PublicStatsTime;
  active_scope: PublicStatsScope;
  filters: PublicStatsFilters;
  regions: PublicStatsRegion[];
  top_issues: PublicTopIssue[];
  generated_at: string;
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
