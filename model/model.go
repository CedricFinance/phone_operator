package model

import "time"

type ForwardingRequest struct {
	Id            string
	RequesterId   string
	RequesterName string
	Duration      int
	CreatedAt     time.Time
	AcceptedAt    *time.Time
	RefusedAt     *time.Time
	ExpiresAt     *time.Time
	AnsweredBy    string
}
