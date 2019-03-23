package main

import (
	"flag"
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/golang/glog"
	"github.com/golang/protobuf/jsonpb"
	bdaybot "github.com/j-fuentes/bdaybot-slack/api"
	"github.com/j-fuentes/bdaybot-slack/internal/calendar/googlesheet"
	"github.com/j-fuentes/bdaybot-slack/internal/oauth2"
	"github.com/j-fuentes/bdaybot-slack/internal/slack"
)

var configFile string
var userTokenFile string
var auth bool

// SheetIDFromURL is a regexp to extract Google sheetID from url
var SheetIDFromURL = regexp.MustCompile(`https://docs.google.com/spreadsheets/d/(?P<id>.*)/.*`)

func init() {
	flag.StringVar(&configFile, "config", "./config.json", "Path to the config file.")
	flag.StringVar(&userTokenFile, "userTokenFile", "./token.json", "Path to a file with user oauth2 token.")
	flag.BoolVar(&auth, "auth", false, "Triggers the auth workflow.")
	flag.Parse()
}

func main() {
	configReader, err := os.Open(configFile)
	if err != nil {
		glog.Fatalf("%+v", err)
	}

	var config bdaybot.Config
	jsonpb.Unmarshal(configReader, &config)

	adminWebhookURL := config.GetSlack().GetAdminWebhookUrl()

	if auth {
		glog.Info("Starting auth workflow...")
		token, err := oauth2.RetrieveTokenInteractively(googlesheet.GenOauthConfig(
			config.GetOauth2().GetClientId(),
			config.GetOauth2().GetClientSecret(),
		))
		if err != nil {
			glog.Fatalf("%v", err)
		}

		err = oauth2.WriteTokenToFile(token, userTokenFile)
		if err != nil {
			glog.Fatalf("%v", err)
		}
		os.Exit(0)
	}

	var slackAdminClient *slack.Client
	if adminWebhookURL != "" {
		glog.Infof("Found slack webhook for admin, events will be reported to %s\n", adminWebhookURL)
		slackAdminClient = slack.NewClient(adminWebhookURL)
	} else {
		glog.Infof("Did not found a slack webhook for admin\n")
	}

	reader, err := genGooglesheetReader(config)
	if err != nil {
		sendErrorToSlack(
			slackAdminClient,
			"Cannot connect to Google Spreadsheets",
			err,
		)
		glog.Fatalf("%+v", err)
	}

	spreadsheetURL := config.GetCalendar().GetGoogleSheet().GetUrl()
	glog.Infof("Reading bdays from %s\n", spreadsheetURL)
	bdays, err := reader.GetBdays()
	if err != nil {
		sendErrorToSlack(
			slackAdminClient,
			fmt.Sprintf("Error while reading data from spreadsheet (%s)", spreadsheetURL),
			err,
		)
		glog.Fatalf("%+v", err)
	}

	glog.Infof("%d names were loaded.\n", len(bdays))

	webhookURL := config.GetSlack().GetWebhookUrl()
	prefix := config.GetSlack().GetSalutePrefix()
	suffix := config.GetSlack().GetSaluteSuffix()

	_, month, day := time.Now().Date()
	names := []string{}
	for _, bday := range bdays {
		if bday.GetDate().GetMonth() == int32(month) &&
			bday.GetDate().GetDay() == int32(day) {
			glog.Infof("It is @%s bday!\n", bday.GetUserId())
			names = append(names, bday.GetUserId())
		}
	}

	if numNames := len(names); numNames > 0 {
		people := fmt.Sprintf("@%s", names[0])
		if numNames > 1 {
			people = fmt.Sprintf("@%s and @%s", strings.Join(names[:len(names)-1], ", "), names[len(names)-1])
		}
		glog.Infof("Sending salute to %s via Slack webhook %s\n", people, webhookURL)
		sendInfoToSlack(
			slackAdminClient,
			fmt.Sprintf("Saying happy birthday to %s", people),
			fmt.Sprintf("%d bdays were found in the spreadsheet (%s)", numNames, spreadsheetURL),
		)
		slackClient := slack.NewClient(webhookURL)
		err := slackClient.SendMessage(fmt.Sprintf("%s %s %s", prefix, people, suffix))
		if err != nil {
			glog.Fatalf("%+v", err)
		}
	} else {
		sendInfoToSlack(
			slackAdminClient,
			"Today is no one's birthday",
			fmt.Sprintf("No birthdays found on the spreadsheet (%s)", spreadsheetURL),
		)
		glog.Info("Today is no one's birthday")
	}
}

func genGooglesheetReader(config bdaybot.Config) (*googlesheet.Reader, error) {
	oauthConfig := googlesheet.GenOauthConfig(config.GetOauth2().GetClientId(), config.GetOauth2().GetClientSecret())

	sheetID, err := getSheetID(config.GetCalendar().GetGoogleSheet())
	if err != nil {
		return nil, fmt.Errorf("%+v", err)
	}

	token, err := oauth2.ReadTokenFromFile(userTokenFile)
	if err != nil {
		return nil, err
	}

	client := oauth2.GetHTTPClient(oauthConfig, token)

	return googlesheet.NewReader(client, sheetID)
}

func getSheetID(sheet *bdaybot.GoogleSheet) (string, error) {
	url := sheet.GetUrl()
	match := SheetIDFromURL.FindStringSubmatch(url)
	if len(match) != 2 {
		return "", fmt.Errorf("Cannot get sheet ID from url %q", url)
	}

	return match[1], nil
}

func sendErrorToSlack(client *slack.Client, title string, err error) {
	if client == nil {
		return
	}

	err2 := client.Send(&slack.SlackRequestBody{
		Text:  "Error from bdaybot",
		Parse: "full",
		Attachments: []*slack.Attachment{
			&slack.Attachment{
				Color:   "#e23434",
				Pretext: "Something bad happened",
				Title:   title,
				Text:    fmt.Sprintf("```%+v```", err),
			},
		},
	})

	if err2 != nil {
		glog.Infof("Cannot send information to slack: %+v", err)
	}
}

func sendInfoToSlack(client *slack.Client, title string, msg string) {
	if client == nil {
		return
	}

	err := client.Send(&slack.SlackRequestBody{
		Text:  "Info from bdaybot",
		Parse: "full",
		Attachments: []*slack.Attachment{
			&slack.Attachment{
				Color: "#2159b2",
				Title: title,
				Text:  msg,
			},
		},
	})

	if err != nil {
		glog.Infof("Cannot send information to slack: %+v", err)
	}
}
