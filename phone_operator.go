package main

import (
	"fmt"
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
}

var config Config
var slackClient *slack.Client

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

	slackClient = slack.New(config.Slack.Token)

	http.HandleFunc("/slash", slashCommandHandler)
	http.HandleFunc("/sms", smsHandler)
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
		startSMSForward(w, command.UserID, command.UserName, durationInMinutes)
		return
	}

	if parts[0] == "stop" {
		stopSMSForward(w, command.UserID)
		return
	}

	_, err = fmt.Fprint(w, "Hello, World!")
}

func startSMSForward(w http.ResponseWriter, id string, name string, minutes int) {
	fmt.Fprintf(w, "I'm asking to forward SMS to you for %d minute(s)", minutes)
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

type SMS struct {
	From string
	Body string
}

func smsHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "Failed to decode the message")
		return
	}

	message := SMS{
		Body: r.FormValue("Body"),
		From: r.FormValue("From"),
	}
	log.Printf("%+v", message)

	_, _, err = slackClient.PostMessage(config.Slack.Channel, slack.MsgOptionText(
		fmt.Sprintf("*Message from:* %s\n```\n%s\n```", message.From, message.Body),
		true,
	))
	if err != nil {
		log.Printf("Error: %v", err)
	}

	fmt.Fprintf(w, "ok")
}
