package messages

import (
	"fmt"
	"github.com/CedricFinance/phone_operator/model"
	"github.com/slack-go/slack"
	"time"
)

func HomeMessage(requests []*model.ForwardingRequest) slack.Message {
	blocks := make([]slack.Block, 2*len(requests)+2)

	blocks[0] = slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.MarkdownType,
			"*Your Requests*",
			false,
			false,
		),
		nil,
		nil,
	)
	blocks[1] = slack.NewDividerBlock()

	for i, request := range requests {
		blocks[2*i+2] = slack.NewSectionBlock(
			slack.NewTextBlockObject(
				slack.MarkdownType,
				getMessage(request),
				false,
				false,
			),
			nil,
			getAccessory(request),
		)
		blocks[2*i+3] = slack.NewDividerBlock()
	}

	return slack.NewBlockMessage(blocks...)
}

func getMessage(request *model.ForwardingRequest) string {
	if request.AcceptedAt != nil {
		return fmt.Sprintf(
			"<!date^%d^{date_short_pretty} {time}|%s> You'll receive text messages until <!date^%d^{date_short_pretty} {time}|%s>\nStatus: %s",
			request.CreatedAt.Unix(),
			request.CreatedAt.String(),
			request.ExpiresAt.Unix(),
			request.ExpiresAt.String(),
			getStatus(request),
		)
	}

	return fmt.Sprintf(
		"<!date^%d^{date_short_pretty} {time}|%s> You asked to receive text messages for %d minute(s)\nStatus: %s",
		request.CreatedAt.Unix(),
		request.CreatedAt.String(),
		request.Duration,
		getStatus(request),
	)
}

func getAccessory(request *model.ForwardingRequest) *slack.Accessory {
	if request.AcceptedAt != nil && request.ExpiresAt.After(time.Now().UTC()) {
		return slack.NewAccessory(
			slack.NewButtonBlockElement(
				"stop", request.Id, slack.NewTextBlockObject(slack.PlainTextType, "Stop", false, false),
			).WithStyle(slack.StyleDanger),
		)
	}
	return nil
}

func getStatus(request *model.ForwardingRequest) interface{} {
	if request.AcceptedAt != nil {
		return fmt.Sprintf(":thumbsup: Accepted by <@%s>", request.AnsweredBy)
	}

	if request.RefusedAt != nil {
		return fmt.Sprintf(":thumbsdown: Refused by <@%s>", request.AnsweredBy)
	}

	return ":question: Waiting for admin's answer"
}
