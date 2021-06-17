package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/CedricFinance/phone_operator/database"
	"github.com/CedricFinance/phone_operator/messages"
	"github.com/CedricFinance/phone_operator/model"
	"github.com/CedricFinance/phone_operator/repository"
	"github.com/slack-go/slack"
	"gopkg.in/yaml.v3"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type Config struct {
	Slack struct {
		VerificationToken string `yaml:"verification_token"`
		Token             string
		Channel           string
	}
	Database struct {
		User     string
		Password string
		Name     string
		Host     string
	}
}

var config Config
var slackClient *slack.Client
var repo *repository.Repository

func main() {

	f, err := os.Open("config.yaml")
	if err != nil {
		panic(fmt.Errorf("failed to load config: %s", err))
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	err = decoder.Decode(&config)
	if err != nil {
		panic(fmt.Errorf("failed to unmarshal config: %s", err))
	}

	slackClient = slack.New(config.Slack.Token, slack.OptionDebug(true))

	db := database.Connect(
		config.Database.User,
		config.Database.User,
		config.Database.Name,
		config.Database.Host,
	)

	repo = repository.New(db)

	http.HandleFunc("/slash", slashCommandHandler)
	http.HandleFunc("/sms", smsHandler)
	http.HandleFunc("/interactivity", interactivityHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
		log.Printf("Defaulting to port %s", port)
	}

	log.Printf("Listening on port %s", port)
	log.Printf("Open http://localhost:%s in the browser", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

func slashCommandHandler(w http.ResponseWriter, r *http.Request) {
	command, err := slack.SlashCommandParse(r)
	if err != nil {
		log.Printf("Failed to parse slash command: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !command.ValidateToken(config.Slack.VerificationToken) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	parts := strings.SplitN(command.Text, " ", 2)

	if parts[0] == "" || parts[0] == "help" {
		showHelp(w)
		return
	}

	if parts[0] == "start" {
		durationInMinutes := 60
		if len(parts) > 1 {
			durationInMinutes, err = parseDuration(parts[1])
			if err != nil {
				fmt.Fprintf(w, err.Error())
				return
			}
		}
		startSMSForward(r.Context(), w, command.UserID, command.UserName, durationInMinutes)
		return
	}

	if parts[0] == "stop" {
		stopSMSForward(w, command.UserID)
		return
	}

	_, err = fmt.Fprint(w, "Hello, World!")
}

func smsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to decode the message")
		return
	}

	message := model.SMS{
		Body: r.FormValue("Body"),
		From: r.FormValue("From"),
	}
	log.Printf("%+v", message)

	activeRequests, _ := repo.GetActiveForwardingRequests(r.Context())
	fmt.Printf("%d active requests", len(activeRequests))

	for _, request := range activeRequests {
		notifyUserBlock(r.Context(), request.RequesterId, messages.SmsUserNotifyMessage(message))
	}

	_, _, err = slackClient.PostMessage(
		config.Slack.Channel,
		slack.MsgOptionBlocks(messages.SmsChannelNotifyMessage(message, activeRequests).Blocks.BlockSet...),
	)
	if err != nil {
		log.Printf("Error: %v", err)
	}

	fmt.Fprintf(w, "ok")
}

func interactivityHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	payload := r.Form.Get("payload")

	var message slack.InteractionCallback
	if err := json.Unmarshal([]byte(payload), &message); err != nil {
		fmt.Printf("Failed to unmarshal %q: %v\n", payload, err)
		fmt.Fprintf(w, err.Error())
		return
	}

	switch message.Type {
	case slack.InteractionTypeBlockActions:
		fmt.Println("block actions")

		if len(message.ActionCallback.BlockActions) > 1 {
			fmt.Println("Received multiple block actions")
			return
		}

		if message.View.CallbackID == "" {
			handleActionFromBlockId(message, r, w)
		}

		return
	}
}

func handleActionFromBlockId(message slack.InteractionCallback, r *http.Request, w http.ResponseWriter) {
	action := message.ActionCallback.BlockActions[0]

	switch action.BlockID {
	case "forwarding_request":
		handleForwardingRequestActions(message, r, w)
	}
}

func handleForwardingRequestActions(message slack.InteractionCallback, r *http.Request, w http.ResponseWriter) {
	action := message.ActionCallback.BlockActions[0].ActionID
	requestId := message.ActionCallback.BlockActions[0].Value

	if action == "accept" {
		acceptForwardingRequest(r.Context(), message, requestId)
	} else {
		refuseForwardingRequest(r.Context(), message, requestId)
	}

}

func refuseForwardingRequest(ctx context.Context, message slack.InteractionCallback, requestId string) {
	repo.RefuseForwardingRequest(ctx, requestId, message.User.ID)
	request, _ := repo.GetForwardingRequest(ctx, requestId)
	notifyUser(
		ctx,
		request.RequesterId,
		"Sorry, your request has been refused.",
	)
	slackClient.PostMessage(
		message.Channel.GroupConversation.Conversation.ID,
		slack.MsgOptionBlocks(messages.AcceptRefuseRequestMessage(request).Blocks.BlockSet...),
		slack.MsgOptionReplaceOriginal(message.ResponseURL),
	)
}

func acceptForwardingRequest(ctx context.Context, message slack.InteractionCallback, requestId string) {
	repo.AcceptForwardingRequest(ctx, requestId, message.User.ID)
	request, _ := repo.GetForwardingRequest(ctx, requestId)
	notifyUser(
		ctx,
		request.RequesterId,
		fmt.Sprintf("Your request has been accepted. I'll forward you the messages until %s", request.ExpiresAt.String()),
	)
	slackClient.PostMessage(
		message.Channel.GroupConversation.Conversation.ID,
		slack.MsgOptionBlocks(messages.AcceptRefuseRequestMessage(request).Blocks.BlockSet...),
		slack.MsgOptionReplaceOriginal(message.ResponseURL),
	)
}

func notifyUser(ctx context.Context, slackId string, message string) error {
	c, _, _, err := slackClient.OpenConversationContext(ctx, &slack.OpenConversationParameters{
		ReturnIM: true,
		Users:    []string{slackId},
	})
	if err != nil {
		return err
	}

	_, _, _, err = slackClient.SendMessage(
		c.ID,
		slack.MsgOptionText(message, false),
	)
	if err != nil {
		return err
	}

	return nil
}

func notifyUserBlock(ctx context.Context, slackId string, message slack.Message) error {
	c, _, _, err := slackClient.OpenConversationContext(ctx, &slack.OpenConversationParameters{
		ReturnIM: true,
		Users:    []string{slackId},
	})
	if err != nil {
		return err
	}

	_, _, _, err = slackClient.SendMessage(
		c.ID,
		slack.MsgOptionBlocks(message.Blocks.BlockSet...),
	)
	if err != nil {
		return err
	}

	return nil
}

func startSMSForward(context context.Context, w http.ResponseWriter, userId string, userName string, duration int) {
	request := repository.NewForwardingRequest(userId, userName, duration)
	err := repo.SaveForwardingRequest(context, request)
	if err != nil {
		fmt.Fprintf(w, "Oops. Something went wrong :sad:. Error: %s", err)
		return
	}

	_, _, err = slackClient.PostMessage(
		config.Slack.Channel,
		slack.MsgOptionBlocks(messages.AcceptRefuseRequestMessage(request).Blocks.BlockSet...),
	)
	if err != nil {
		fmt.Fprintf(w, "Oops. Something went wrong :sad:. Error: %s", err)
		return
	}

	fmt.Fprintf(w, "I'm asking to forward SMS to you for %d minute(s)", duration)
}

func stopSMSForward(w http.ResponseWriter, id string) {
	fmt.Fprintf(w, "Ok, I won't forward you more SMS")
}

func parseDuration(durationStr string) (int, error) {
	pattern := regexp.MustCompile("([0-9])+\\s*([a-zA-Z]*)")
	result := pattern.FindStringSubmatch(durationStr)
	fmt.Printf("%v", result)

	if len(result) == 0 {
		return 0, fmt.Errorf("I don't understand the duration you want. Please enter a number followed by a unit:\n- `m`, `min`, `minute`, `minutes` for minutes\n- `h`, `hour`, `hours` for hours\n- `d`, `day`, `days` for days\n\nNote: you can omit the unit for minutes")
	}

	var err error
	duration := 0
	minutesPerUnit := 1

	if len(result) > 1 {
		duration64, err := strconv.ParseInt(result[1], 10, 32)
		duration = int(duration64)
		if err != nil {
			return 0, fmt.Errorf("I don't understand the duration you want. %q is not a number.", result[1])
		}
	}

	if len(result) > 2 {
		minutesPerUnit, err = getMinutesPerUnit(result[2])
		if err != nil {
			return 0, fmt.Errorf("I don't understand the duration you want. %q is not a valid unit. Please use:\n- `m`, `min`, `minute`, `minutes` for minutes\n- `h`, `hour`, `hours` for hours\n- `d`, `day`, `days` for days\n", result[2])
		}
	}
	return duration * minutesPerUnit, nil
}

func getMinutesPerUnit(unit string) (int, error) {
	switch unit {
	case "m", "min", "minute", "minutes":
		return 1, nil
	case "h", "hour", "hours":
		return 60, nil
	case "d", "day", "days":
		return 24 * 60, nil
	}

	return 0, fmt.Errorf("%q is not a valid unit", unit)
}

func showHelp(w http.ResponseWriter) {
	fmt.Fprintf(w, "Available commands:\n`/sms help` - display this help message\n`/sms start [duration]` - ask to start SMS forwarding for [duration] (default duration is 1h)\n`/sms stop` - stop SMS forwarding")
}
