package main

import (
    "net/http"
    "strings"
    "testing"
)

func TestParseNexmoSMS(t *testing.T) {
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
