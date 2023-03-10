package main

import (
    "context"
    "github.com/CedricFinance/phone_operator/model"
    "io"
    "net/http"
    "net/http/httptest"
    "testing"
)

func parseSMS(_ *http.Request) (model.SMS, error) {
    return model.SMS{
        From: "0123456789",
        Body: "HelloWorld",
    }, nil
}

func TestSMSHandler_ServeHTTP(t *testing.T) {
    var handler = WebhookHandler[model.SMS]{
        Parser: parseSMS,
        Handler: func(ctx context.Context, message model.SMS) error {
            if message.From != "0123456789" {
                t.Errorf("invalid message.From, expected: %q, got: %q", "0123456789", message.From)
            }

            if message.Body != "HelloWorld" {
                t.Errorf("invalid message.Body, expected: %q, got: %q", "HelloWorld", message.Body)
            }

            return nil
        },
    }

    r, _ := http.NewRequest("POST", "http://localhost", nil)
    w := httptest.NewRecorder()

    handler.ServeHTTP(w, r)

    res := w.Result()
    if res.StatusCode != 200 {
        t.Errorf("Expected HTTP Code 200, got: %d", res.StatusCode)
    }

    body, _ := io.ReadAll(res.Body)
    if string(body) != "" {
        t.Errorf("Expected empty body, got: %q", body)
    }

}
