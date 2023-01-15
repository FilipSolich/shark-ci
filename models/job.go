package models

type Job struct {
	ID              string      `json:"_id,omitempty" bson:"_id,omitempty"`
	RepoID          string      `json:"-" bson:"repo,omitempty"`
	Token           OAuth2Token `json:"token,omitempty" bson:"token,omitempty"`
	CommitSHA       string      `json:"commmitSHA,omitempty" bson:"commmitSHA,omitempty"`
	CloneURL        string      `json:"cloneURL,omitempty" bson:"cloneURL,omitempty"`
	TargetURL       string      `json:"targetURL,omitempty" bson:"targetURL,omitempty"`
	ReportStatusURL string      `json:"reportStatusURL,omitempty" bson:"reportStatusURL,omitempty"`
	PublishLogsURL  string      `json:"publishLogsURL,omitempty" bson:"publishLogsURL,omitempty"`
}
