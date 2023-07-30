package models

// TODO: Add owner.

type Repo struct {
	ID            int64  `json:"id"`
	Service       string `json:"service"`
	RepoServiceID int64  `json:"repo_service_id"`
	Name          string `json:"name"`
	WebhookID     *int64 `json:"webhook_id,omitempty"`
	ServiceUserID int64  `json:"service_user_id"`
}
