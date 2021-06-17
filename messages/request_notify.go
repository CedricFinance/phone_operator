package messages

import (
	"fmt"
	"github.com/CedricFinance/phone_operator/model"
	"github.com/slack-go/slack"
)

func AcceptedRequestMessage(request *model.ForwardingRequest) slack.Message {
	return slack.NewBlockMessage(
		slack.NewSectionBlock(
			slack.NewTextBlockObject(
				slack.MarkdownType,
				fmt.Sprintf("Your request has been accepted. I'll forward you the messages until <!date^%d^{date_short_pretty} {time}|%s>", request.ExpiresAt.Unix()),
				false,
				false,
			),
			nil,
			nil,
		),
	)
}
