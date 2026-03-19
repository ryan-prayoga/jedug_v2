package domain

import "time"

type IssueFollowState struct {
	Following                    bool       `json:"following"`
	FollowersCount               int        `json:"followers_count"`
	FollowerToken                string     `json:"follower_token,omitempty"`
	FollowerTokenExpiresAt       *time.Time `json:"follower_token_expires_at,omitempty"`
	FollowerStreamToken          string     `json:"follower_stream_token,omitempty"`
	FollowerStreamTokenExpiresAt *time.Time `json:"follower_stream_token_expires_at,omitempty"`
}

type IssueFollowersCount struct {
	FollowersCount int `json:"followers_count"`
}
