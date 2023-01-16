package models

type Repo struct {
	ID            string `bson:"_id,omitempty"`
	RepoServiceID int64  `bson:"repoServiceID,omitempty"`
	ServiceName   string `bson:"serviceName"`
	Name          string `bson:"name,omitempty"`
	FullName      string `bson:"fullName,omitempty"`
	UniqueName    string `bson:"uniqueName"`
	WebhookID     int64  `bson:"webhookID,omitempty"`
	WebhookActive bool   `bson:"webhookActive,omitempty"`
}

func NewRepo(repoID int64, serviceName string, name string, fullName string) *Repo {
	return &Repo{
		RepoServiceID: repoID,
		ServiceName:   serviceName,
		Name:          name,
		FullName:      fullName,
		UniqueName:    serviceName + "/" + fullName,
	}
}

func (r *Repo) DeleteWebhook() {
	r.WebhookID = 0
	r.WebhookActive = false
}
