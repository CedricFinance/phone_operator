package messages

import (
	"fmt"
	"github.com/CedricFinance/phone_operator/model"
	"github.com/slack-go/slack"
	"strings"
)

func SmsChannelNotifyMessage(message model.SMS, activeRequests []*model.ForwardingRequest) slack.Message {
	usersList := generateUserReferences(activeRequests)

	return slack.NewBlockMessage(
		smsMessageBlock(message),
		slack.NewContextBlock(
			"context",
			slack.NewTextBlockObject(
				slack.MarkdownType,
				fmt.Sprintf(":incoming_envelope: Forwarded to %d user(s): %s", len(activeRequests), usersList),
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

func generateUserReferences(requests []*model.ForwardingRequest) interface{} {
	references := make([]string, len(requests))

	for i, request := range requests {
		references[i] = fmt.Sprintf("<@%s>", request.RequesterId)
	}

	return strings.Join(references, ", ")
}
