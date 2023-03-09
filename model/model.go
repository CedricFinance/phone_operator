package model

import (
    "net/http"
    "time"
)

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

func (r ForwardingRequest) IsActive() bool {
    return r.ExpiresAt != nil && r.ExpiresAt.After(time.Now().UTC())
}

func (r ForwardingRequest) IsPending() bool {
    return r.AcceptedAt == nil && r.RefusedAt == nil
}

type SMS struct {
    From string
    Body string
}

type SMSParser func(r *http.Request) (SMS, error)
