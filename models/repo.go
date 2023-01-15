package models

type Repo struct {
	ID            string `bson:"_id,omitempty"`
	RepoServiceID int64  `bson:"repoServiceID,omitempty"`
	ServiceName   string `bson:"serviceName"`
	Name          string `bson:"name,omitempty"`
	FullName      string `bson:"fullName,omitempty"`
	UniqueName    string `bson:"uniqueName"`
	WebhookID     int64  `bson:"webhookID,omitempty"`
	WebhookActive bool   `bson:"active,omitempty"`
}
