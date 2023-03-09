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
