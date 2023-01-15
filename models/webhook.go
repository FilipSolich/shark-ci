package models

type Webhook struct {
	WebhookID int64 `bson:"webhookID,omitempty"`
	Active    bool  `bson:"active,omitempty"`
}
