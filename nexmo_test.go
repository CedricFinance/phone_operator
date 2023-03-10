package main

import (
    "net/http"
    "strings"
    "testing"
)

func TestParseNexmoSMS_POST(t *testing.T) {
    r, _ := http.NewRequest(http.MethodPost, "http://localhost", strings.NewReader("{ \"text\": \"HelloWorld\", \"to\": \"0612345678\", \"msisdn\": \"0123456789\", \"type\": \"text\" }"))
    r.Header.Set("Content-Type", "application/json")

    message, err := ParseNexmoSMS(r)
    if err != nil {
        t.Errorf("failed to parse incoming SMS")
    }

    if message.Body != "HelloWorld" {
        t.Errorf("invalid Body, expected: %q, got: %q", "HelloWorld", message.Body)
    }

    if message.From != "0123456789" {
        t.Errorf("invalid From, expected: %q, got: %q", "0123456789", message.From)
    }
}

func TestParseNexmoSMS_GET(t *testing.T) {
    r, _ := http.NewRequest(http.MethodGet, "http://localhost?text=HelloWorld&msisdn=0123456789", nil)
    r.Header.Set("Content-Type", "application/json")

    message, err := ParseNexmoSMS(r)
    if err != nil {
        t.Errorf("failed to parse incoming SMS")
    }

    if message.Body != "HelloWorld" {
        t.Errorf("invalid Body, expected: %q, got: %q", "HelloWorld", message.Body)
    }

    if message.From != "0123456789" {
        t.Errorf("invalid From, expected: %q, got: %q", "0123456789", message.From)
    }
}

func TestParseNexmoPhoneCallEvent_GET(t *testing.T) {
    r, _ := http.NewRequest(http.MethodGet, "http://localhost?status=ok&from=0123456789&call_duration=9&call_start=2023-03-10+16%3A49%3A43&call_end=2023-03-10+16%3A50%3A17", nil)
    r.Header.Set("Content-Type", "application/json")

    message, err := ParseNexmoPhoneCallEvent(r)
    if err != nil {
        t.Errorf("failed to parse incoming Phone call event")
    }

    if message.Status != "ok" {
        t.Errorf("invalid Status, expected: %q, got: %q", "ok", message.Status)
    }

    if message.From != "0123456789" {
        t.Errorf("invalid From, expected: %q, got: %q", "0123456789", message.From)
    }

    if message.Start != "2023-03-10 16:49:43" {
        t.Errorf("invalid Start, expected: %q, got: %q", "2023-03-10 16:49:43", message.Start)
    }

    if message.End != "2023-03-10 16:50:17" {
        t.Errorf("invalid End, expected: %q, got: %q", "2023-03-10 16:50:17", message.End)
    }

}
