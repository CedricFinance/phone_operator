package main

import (
    "encoding/json"
    "fmt"
    "github.com/CedricFinance/phone_operator/model"
    "net/http"
)

type nexmoIncomingSms struct {
    Type string `json:"type"`
    From string `json:"msisdn"`
    To   string `json:"to"`
    Text string `json:"text"`
}

func ParseNexmoSMS(r *http.Request) (model.SMS, error) {
    var incomingSms nexmoIncomingSms

    if r.Method == http.MethodPost {
        err := json.NewDecoder(r.Body).Decode(&incomingSms)
        if err != nil {
            return model.SMS{}, fmt.Errorf("failed to parse request body: %w", err)
        }

        message := model.SMS{
            Body: incomingSms.Text,
            From: incomingSms.From,
        }

        return message, nil
    }

    if r.Method == http.MethodGet {
        message := model.SMS{
            Body: r.URL.Query().Get("text"),
            From: r.URL.Query().Get("msisdn"),
        }

        return message, nil
    }

    return model.SMS{}, fmt.Errorf("can't handle %q HTTP method", r.Method)
}
