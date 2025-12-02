package model

import "time"

type MessageStatus string

const (
	StatusPending MessageStatus = "PENDING"
	StatusSent    MessageStatus = "SENT"
	StatusFailed  MessageStatus = "FAILED"
)

type Message struct {
	ID           int64         `db:"id" json:"id"`
	PhoneNumber  string        `db:"phone_number" json:"phone_number"`
	Content      string        `db:"content" json:"content"`
	Status       MessageStatus `db:"status" json:"status"`
	SentTime     time.Time     `db:"sent_time" json:"sent_time"`
	ExternalID   string        `db:"external_id" json:"external_id"`
	AttemptCount int           `db:"attempt_count" json:"attempt_count"`
}
