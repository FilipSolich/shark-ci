package models

type Repo struct {
	ID          string  `bson:"_id,omitempty"`
	RepoID      int64   `bson:"repoID,omitempty"`
	ServiceName string  `bson:"serviceName"`
	Name        string  `bson:"name,omitempty"`
	FullName    string  `bson:"fullName,omitempty"`
	Webhook     Webhook `bson:"webhook,omitempty"`
}
