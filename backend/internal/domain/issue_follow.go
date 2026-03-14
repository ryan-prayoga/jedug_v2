package domain

type IssueFollowState struct {
	Following      bool `json:"following"`
	FollowersCount int  `json:"followers_count"`
}

type IssueFollowersCount struct {
	FollowersCount int `json:"followers_count"`
}
