package messages

import (
	"fmt"
	"github.com/CedricFinance/phone_operator/model"
	"github.com/slack-go/slack"
)

func AcceptRefuseRequestMessage(request *model.ForwardingRequest) slack.Message {
	return slack.NewBlockMessage(
		acceptRefuseMessageBlock(request),
		acceptRefuseActionsBlock(request),
	)
}

func acceptRefuseMessageBlock(request *model.ForwardingRequest) *slack.SectionBlock {
	return slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.MarkdownType,
			fmt.Sprintf("<@%s> want's to receive SMS for %d minute(s)", request.RequesterId, request.Duration),
			false,
			false,
		),
		nil,
		nil,
	)
}

func acceptRefuseActionsBlock(request *model.ForwardingRequest) slack.Block {
	if request.AcceptedAt != nil {
		return slack.NewContextBlock(
			"accepted",
			slack.NewTextBlockObject(
				slack.MarkdownType,
				fmt.Sprintf(
					":thumbsup: Accepted <!date^%d^{date_short_pretty}|%s> by <@%s>",
					request.AcceptedAt.Unix(),
					request.AcceptedAt.String(),
					request.AnsweredBy,
				),
				false,
				false,
			),
		)
	}

	if request.RefusedAt != nil {
		return slack.NewContextBlock(
			"refused",
			slack.NewTextBlockObject(
				slack.MarkdownType,
				fmt.Sprintf(
					":thumbsdown: Refused <!date^%d^{date_short_pretty}|%s> by <@%s>",
					request.RefusedAt.Unix(),
					request.RefusedAt.String(),
					request.AnsweredBy,
				),
				false,
				false,
			),
		)
	}

	return slack.NewActionBlock(
		"forwarding_request",
		slack.NewButtonBlockElement(
			"accept",
			request.Id,
			slack.NewTextBlockObject(slack.PlainTextType, ":thumbsup: Accept", false, false)).WithStyle(slack.StylePrimary),
		slack.NewButtonBlockElement(
			"refuse",
			request.Id,
			slack.NewTextBlockObject(slack.PlainTextType, ":thumbsdown: Refuse", false, false)).WithStyle(slack.StyleDanger),
	)
}
