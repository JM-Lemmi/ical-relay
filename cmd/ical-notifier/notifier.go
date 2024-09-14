package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"rss"
	"time"

	"gopkg.in/gomail.v2"

	ics "github.com/arran4/golang-ical"
	"github.com/jm-lemmi/ical-relay/compare"
	"github.com/jm-lemmi/ical-relay/datastore"
	"github.com/jm-lemmi/ical-relay/helpers"
	log "github.com/sirupsen/logrus"
)

func notifyChanges(notifierName string, notifier datastore.Notifier) error {
	requestLogger := log.WithFields(log.Fields{"notifier": notifierName})
	requestLogger.Infoln("Running Notifier!")
	requestLogger.Tracef("Notifier: %v", notifier)

	// get source
	currentICS, err := getSource(notifier.Source)
	if err != nil {
		log.Error("Failed to get source for notifier", notifierName, err)
		return err
	}
	// compare to history on file
	historyFilename := conf.General.StoragePath + "notifystore/" + notifierName + "-past.ics"
	historyICS, err := helpers.LoadCalFile(historyFilename)

	// if history file does not exist, create it
	if err != nil {
		log.Info("History file does not exist, saving for the first time")
		helpers.WriteCalFile(currentICS, historyFilename)
		return nil
	}

	added, deleted, changed_old, changed_new := compare.Compare(historyICS, currentICS)

	if len(added)+len(deleted)+len(changed_new) == 0 {
		log.Info("No changes detected.")
		return nil
	}
	log.Debug("Changes detected: " + fmt.Sprint(len(added)) + " added, " + fmt.Sprint(len(deleted)) + " deleted, " + fmt.Sprint(len(changed_new)) + " changed")

	// iterate over all recipients by type
	errcnt := 0
	for _, rec := range notifier.Recipients {
		log.Debug("notifier " + rec.Type + " " + rec.Recipient)
		var err error
		switch rec.Type {
		case "mail":
			err = sendNotifyMail(notifierName, rec.Recipient, added, deleted, changed_new)

		case "rss":
			err = sendRSSFeed(notifierName, rec.Recipient, added, deleted, changed_new)

		case "database":
			err = sendDatabaseHistory(notifierName, rec.Recipient, added, deleted, changed_old, changed_new)

		case "webhook":
			err = sendNotifyWebhook(notifierName, rec.Recipient, added, deleted, changed_new)
		}

		if err != nil {
			log.Errorf("Failed to devliver notifier %s for recipient %s: %v", notifierName, rec.Recipient, err)
			errcnt++
		}
	}

	// save updated calendar
	log.Debugf("Saving Updated Calendar to %s", historyFilename)
	helpers.WriteCalFile(currentICS, historyFilename)

	if errcnt != 0 {
		// at least one error occured
		return fmt.Errorf("%d of %d notifier recipients failed to run", errcnt, len(notifier.Recipients))
	} else {
		return nil
	}

}

func sendNotifyMail(notifierName string, recipient string, added []ics.VEvent, deleted []ics.VEvent, changed []ics.VEvent) error {
	var body string

	if len(added) > 0 {
		for _, event := range added {
			body += "Added:\n\n" + helpers.PrettyPrint(event) + "\n\n"
		}
	}
	if len(deleted) > 0 {
		for _, event := range deleted {
			body += "Deleted:\n\n" + helpers.PrettyPrint(event) + "\n\n"
		}
	}
	if len(changed) > 0 {
		for _, event := range changed {
			body += "Changed (displaying new version):\n\n" + helpers.PrettyPrint(event) + "\n\n"
		}
	}

	m := gomail.NewMessage()
	m.SetHeader("From", conf.General.Mail.Sender)
	m.SetHeader("To", recipient)
	m.SetHeader("Subject", "Calendar Notification for "+notifierName)

	if !conf.General.LiteMode {
		unsubscribeURL := conf.General.URL + "/view/" + url.QueryEscape(notifierName) + "/unsubscribe?type=mail&recipient=" + url.QueryEscape(recipient)
		m.SetHeader("List-Unsubscribe", unsubscribeURL)
		body += "\n\nUnsubscribe: " + unsubscribeURL
	}

	m.SetBody("text/plain", body)

	var d gomail.Dialer
	if conf.General.Mail.SMTPUser != "" && conf.General.Mail.SMTPPass != "" {
		d = gomail.Dialer{
			Host:     conf.General.Mail.SMTPServer,
			Port:     conf.General.Mail.SMTPPort,
			Username: conf.General.Mail.SMTPUser,
			Password: conf.General.Mail.SMTPPass,
			SSL:      conf.General.Mail.SMTPSSL,
		}
	} else {
		d = gomail.Dialer{
			Host: conf.General.Mail.SMTPServer,
			Port: conf.General.Mail.SMTPPort,
			SSL:  conf.General.Mail.SMTPSSL,
		}
	}
	log.Info("Sending Mail Notification to " + recipient)

	err := d.DialAndSend(m)
	return err
}

func sendRSSFeed(notifierName string, filename string, added []ics.VEvent, deleted []ics.VEvent, changed []ics.VEvent) error {
	var feed rss.Rss
	var file *os.File
	var err error

	if _, err := os.Stat(filename); errors.Is(err, os.ErrNotExist) {
		log.Info("RSS feed does not exist, creating new one")
		// create new feed
		feed = rss.Rss{
			Title:       "Calendar Notification for " + notifierName,
			Description: "This is a notification feed for changes in the calendar " + notifierName,
			Link:        conf.General.URL + "/view/" + url.QueryEscape(notifierName) + "/changefeed",
			Version:     "2.0",
			Item: []rss.Item{
				{
					Title:       "Feed created",
					Description: "The feed was created. Changes before this date were not recorded.",
					PubDate:     time.Now().Format(time.RFC1123Z),
				},
			},
		}

		// Write the new RSS feed to the file
		file, err = os.Create(filename)
		if err != nil {
			return err
		}

	} else {
		file, err = os.Open(filename)
		if err != nil {
			return err
		}

		// parse existing feed
		feed, err = rss.ParseRSS(file)
		if err != nil {
			return err
		}
	}

	// add new items
	for _, event := range added {
		feed.AddItem(rss.Item{
			Title:       "Added " + event.GetProperty(ics.ComponentPropertySummary).Value,
			Description: helpers.PrettyPrint(event),
			PubDate:     time.Now().Format(time.RFC1123Z),
		})
	}
	for _, event := range deleted {
		feed.AddItem(rss.Item{
			Title:       "Deleted " + event.GetProperty(ics.ComponentPropertySummary).Value,
			Description: helpers.PrettyPrint(event),
			PubDate:     time.Now().Format(time.RFC1123Z),
		})
	}
	for _, event := range changed {
		feed.AddItem(rss.Item{
			Title:       "Changed " + event.GetProperty(ics.ComponentPropertySummary).Value,
			Description: helpers.PrettyPrint(event),
			PubDate:     time.Now().Format(time.RFC1123Z),
		})
	}

	err = feed.WriteRSS(file)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}

	return err
}

type dataObject struct {
	Before ics.VEvent
	After  ics.VEvent
}

func sendDatabaseHistory(notifierName string, recipient string, added []ics.VEvent, deleted []ics.VEvent, changed_old []ics.VEvent, changed_new []ics.VEvent) error {
	changedTime := time.Now()
	for _, event := range added {
		object := &dataObject{After: event}
		jsonB, err := json.Marshal(object)
		if err != nil {
			return err
		}
		eventTime := getTimeFromEvent(event)
		dataStore.AddNotifierHistory(notifierName, recipient, "add", eventTime, changedTime, string(jsonB))
	}
	for _, event := range deleted {
		object := &dataObject{Before: event}
		jsonB, err := json.Marshal(object)
		if err != nil {
			return err
		}
		eventTime := getTimeFromEvent(event)
		dataStore.AddNotifierHistory(notifierName, recipient, "delete", eventTime, changedTime, string(jsonB))
	}
	for index, oldEvent := range changed_old {
		newEvent := changed_new[index]
		if oldEvent.Id() != newEvent.Id() {
			return fmt.Errorf("failed to save notifier history: oldEvent id does not match newEvent id")
		}
		object := &dataObject{Before: oldEvent, After: newEvent}
		jsonB, err := json.Marshal(object)
		if err != nil {
			return err
		}
		eventTime := getTimeFromEvent(oldEvent)
		dataStore.AddNotifierHistory(notifierName, recipient, "change", eventTime, changedTime, string(jsonB))
	}
	return nil
}

// help: how a discord webhook is structured https://gist.github.com/Birdie0/78ee79402a4301b1faf412ab5f1cdcf9
type discordWebhook struct {
	Username string                `json:"username"`
	Avatar   string                `json:"avatar_url"`
	Content  string                `json:"content"`
	Embed    []discordWebhookEmbed `json:"embed"`
}

type discordWebhookEmbed struct {
	// TODO: author object skipped
	Title       string               `json:"title"`
	URL         string               `json:"url"`
	Description string               `json:"description"`
	Color       int                  `json:"color"`
	Footer      discordWebhookFooter `json:"footer"`
	// Thumbnail and Image skipped
}

type discordWebhookFooter struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url"`
}

func sendNotifyWebhook(notifierName string, recipient string, added []ics.VEvent, deleted []ics.VEvent, changed []ics.VEvent) error {
	var embeds []discordWebhookEmbed

	if len(added) > 0 {
		for _, event := range added {
			embeds = append(embeds, discordWebhookEmbed{
				Title:       "Added " + event.GetSummary(),
				Description: helpers.PrettyPrint(event),
				Color:       65280, // color green as int (0x00ff00)
			})
		}
	}
	if len(deleted) > 0 {
		for _, event := range deleted {
			embeds = append(embeds, discordWebhookEmbed{
				Title:       "Deleted " + event.GetSummary(),
				Description: helpers.PrettyPrint(event),
				Color:       16711680, // color red as int (0xff0000)
			})
		}
	}
	if len(changed) > 0 {
		for _, event := range changed {
			embeds = append(embeds, discordWebhookEmbed{
				Title:       "Changed " + event.GetSummary(),
				Description: "(Displaying new version)\n" + helpers.PrettyPrint(event),
				Color:       16744192, // color orange as int (0xff6600)
			})
		}
	}

	webhookBody := discordWebhook{
		Username: "Notifier für " + notifierName,
		Avatar:   conf.General.URL + "static/media/favicon.svg",
		Embed:    embeds,
	}
	webhookBodyJSON, err := json.Marshal(webhookBody)
	if err != nil {
		return fmt.Errorf("error Marshalling webhook json: %v", err)
	}

	log.Info("Sending Discord Webhook Notification to " + recipient)

	req, err := http.NewRequest(http.MethodPost, recipient, bytes.NewBuffer(webhookBodyJSON))
	if err != nil {
		return fmt.Errorf("error creating new Request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")
	req.Header.Set("User-Agent", "Go-http-client/1.1 (ical-notifier/"+version+"; +"+conf.General.URL)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error doing webhook Request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("received status code %d in webhook request: %s", resp.StatusCode, body)
	}
	return err
}

func getTimeFromEvent(event ics.VEvent) time.Time {
	eventTime, err := event.GetStartAt()
	if err != nil {
		eventTime, err = event.GetEndAt()
		if err != nil {
			eventTime = time.Now()
		}
	}
	return eventTime
}
