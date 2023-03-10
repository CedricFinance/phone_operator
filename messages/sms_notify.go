package messages

import (
    "fmt"
    "github.com/CedricFinance/phone_operator/model"
    "github.com/slack-go/slack"
    "strings"
)

func SmsChannelNotifyMessage(message model.SMS, userIds []string) slack.Message {
    usersList := generateUserReferences(userIds)

    return slack.NewBlockMessage(
        smsMessageBlock(message),
        slack.NewContextBlock(
            "context",
            slack.NewTextBlockObject(
                slack.MarkdownType,
                fmt.Sprintf(":incoming_envelope: Forwarded to %d user(s): %s", len(userIds), usersList),
                false,
                false,
            ),
        ),
    )
}

func SmsUserNotifyMessage(message model.SMS) slack.Message {
    return slack.NewBlockMessage(
        smsMessageBlock(message),
    )
}

func smsMessageBlock(message model.SMS) *slack.SectionBlock {
    return slack.NewSectionBlock(
        slack.NewTextBlockObject(
            slack.MarkdownType,
            fmt.Sprintf("*Message from:* %s\n```\n%s\n```", message.From, message.Body),
            false,
            false,
        ),
        nil,
        nil,
    )
}

func generateUserReferences(userIds []string) interface{} {
    references := make([]string, len(userIds))

    for i, userId := range userIds {
        references[i] = fmt.Sprintf("<@%s>", userId)
    }

    return strings.Join(references, ", ")
}

func PhoneCallChannelNotifyMessage(event model.PhoneCallEvent) slack.Message {
    return slack.NewBlockMessage(
        phoneCallMessageBlock(event),
    )
}

func phoneCallMessageBlock(event model.PhoneCallEvent) *slack.SectionBlock {
    return slack.NewSectionBlock(
        slack.NewTextBlockObject(
            slack.MarkdownType,
            fmt.Sprintf("*Phone call from:* %s, received on %s (%ss)\n", event.From, event.Start, event.Duration),
            false,
            false,
        ),
        nil,
        nil,
    )
}
