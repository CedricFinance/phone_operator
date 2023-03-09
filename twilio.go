package main

import (
    "fmt"
    "github.com/CedricFinance/phone_operator/model"
    "net/http"
)

func ParseTwilioSMS(r *http.Request) (model.SMS, error) {
    err := r.ParseForm()
    if err != nil {
        return model.SMS{}, fmt.Errorf("failed to parse request form data: %w", err)
    }

    message := model.SMS{
        Body: r.FormValue("Body"),
        From: r.FormValue("From"),
    }

    return message, nil
}
