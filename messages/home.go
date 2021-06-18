package messages

import (
	"fmt"
	"github.com/CedricFinance/phone_operator/model"
	"github.com/slack-go/slack"
	"time"
)

func HomeMessage(requests []*model.ForwardingRequest) slack.Message {
	var activeRequests []*model.ForwardingRequest
	var pendingRequests []*model.ForwardingRequest
	var pastRequests []*model.ForwardingRequest

	for _, request := range requests {
		if request.IsActive() {
			activeRequests = append(activeRequests, request)
		} else if request.IsPending() {
			pendingRequests = append(pendingRequests, request)
		} else {
			pastRequests = append(pastRequests, request)
		}
	}

	var blocks []slack.Block

	blocks = append(blocks, slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.MarkdownType,
			"*Your active requests*",
			false,
			false,
		),
		nil,
		nil,
	))
	blocks = append(blocks, slack.NewDividerBlock())
	blocks = addRequestsBlocks(activeRequests, blocks)
	blocks = append(blocks, slack.NewContextBlock("", slack.NewImageBlockElement("https://api.slack.com/img/blocks/bkb_template_images/placeholder.png", "placeholder")))

	blocks = append(blocks, slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.MarkdownType,
			"*Your pending requests*",
			false,
			false,
		),
		nil,
		nil,
	))
	blocks = append(blocks, slack.NewDividerBlock())
	blocks = addRequestsBlocks(pendingRequests, blocks)
	blocks = append(blocks, slack.NewContextBlock("", slack.NewImageBlockElement("https://api.slack.com/img/blocks/bkb_template_images/placeholder.png", "placeholder")))

	blocks = append(blocks, slack.NewSectionBlock(
		slack.NewTextBlockObject(
			slack.MarkdownType,
			"*Your past requests*",
			false,
			false,
		),
		nil,
		nil,
	))
	blocks = append(blocks, slack.NewDividerBlock())
	blocks = addRequestsBlocks(pastRequests, blocks)

	return slack.NewBlockMessage(blocks...)
}

func addRequestsBlocks(activeRequests []*model.ForwardingRequest, blocks []slack.Block) []slack.Block {
	for _, request := range activeRequests {
		blocks = append(blocks, slack.NewSectionBlock(
			slack.NewTextBlockObject(
				slack.MarkdownType,
				getMessage(request),
				false,
				false,
			),
			nil,
			getAccessory(request),
		))
		blocks = append(blocks, slack.NewDividerBlock())
	}
	return blocks
}

func getMessage(request *model.ForwardingRequest) string {
	if request.IsActive() {
		return fmt.Sprintf(
			"*<!date^%d^{date_short_pretty} {time}|%s>* You'll receive text messages until <!date^%d^{date_short_pretty} {time}|%s>\n*Status*: %s",
			request.CreatedAt.Unix(),
			request.CreatedAt.String(),
			request.ExpiresAt.Unix(),
			request.ExpiresAt.String(),
			getStatus(request),
		)
	}

	if request.AcceptedAt != nil {
		return fmt.Sprintf(
			"*<!date^%d^{date_short_pretty} {time}|%s>* You receveid text messages from <!date^%d^{date_short_pretty} {time}|%s> to <!date^%d^{date_short_pretty} {time}|%s>\n*Status*: %s",
			request.CreatedAt.Unix(),
			request.CreatedAt.String(),
			request.AcceptedAt.Unix(),
			request.AcceptedAt.String(),
			request.ExpiresAt.Unix(),
			request.ExpiresAt.String(),
			getStatus(request),
		)
	}

	return fmt.Sprintf(
		"*<!date^%d^{date_short_pretty} {time}|%s>* You asked to receive text messages for %d minute(s)\n*Status*: %s",
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
