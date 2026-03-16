import { apiPost } from "./client";
import type { ApiResponse } from "./types";

export interface FollowerAuthResponse {
  follower_id: string;
  follower_token: string;
  expires_at: string;
}

export async function issueFollowerAuthToken(
  followerID: string,
  anonToken: string,
): Promise<ApiResponse<FollowerAuthResponse>> {
  return apiPost<FollowerAuthResponse>(
    "/api/v1/followers/auth",
    { follower_id: followerID },
    anonToken,
  );
}
