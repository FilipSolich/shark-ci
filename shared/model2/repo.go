package model2

type Repo struct {
	ID            int64  `json:"id"`
	RepoServiceID int64  `json:"repo_service_id"`
	Name          string `json:"name"`
	Service       string `json:"service"`
	WebhookID     int64  `json:"webhook_id"`
	WebhookActive bool   `json:"webhook_active"`
	ServiceUserID int64  `json:"service_user_id"`
}
