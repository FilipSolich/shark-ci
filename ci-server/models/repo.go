package models

// TODO: Add owner.

type Repo struct {
	ID            int64  `json:"id"`
	Name          string `json:"name"`
	Owner         string `json:"owner"`
	Service       string `json:"service"`
	RepoServiceID int64  `json:"repo_service_id"`
	WebhookID     *int64 `json:"webhook_id,omitempty"`
	ServiceUserID int64  `json:"service_user_id"`
}
