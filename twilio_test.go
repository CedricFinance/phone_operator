package main

import (
    "net/http"
    "strings"
    "testing"
)

func TestParseTwilioSMS(t *testing.T) {
    r, _ := http.NewRequest(http.MethodPost, "http://localhost", strings.NewReader("From=0123456789&Body=HelloWorld"))
    r.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    message, err := ParseTwilioSMS(r)
    if err != nil {
        t.Errorf("failed to parse incoming SMS")
    }

    if message.Body != "HelloWorld" {
        t.Errorf("invalid Body")
    }

    if message.From != "0123456789" {
        t.Errorf("invalid From")
    }
}
