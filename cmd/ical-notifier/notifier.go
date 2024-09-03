package main

import (
	"encoding/json"
	"errors"
	"fmt"
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

	if len(added)+len(deleted)+len(changed_old) == 0 {
		log.Info("No changes detected.")
		return nil
	}
	log.Debug("Changes detected: " + fmt.Sprint(len(added)) + " added, " + fmt.Sprint(len(deleted)) + " deleted, " + fmt.Sprint(len(changed_old)) + " changed")

	//get notifier including recipients because the method to get all notifiers returns only the notifiers with an empty recipients slice
	notifierInclRep := dataStore.GetNotifier(notifier.Name)
	// iterate over all recipients by type
	for _, rec := range notifierInclRep.Recipients {
		log.Debug("notifier " + rec.Recipient)
		var err error
		switch rec.Type {
		case "mail":
			err = sendNotifyMail(notifierName, rec.Recipient, added, deleted, changed_old)

		case "rss":
			err = sendRSSFeed(notifierName, rec.Recipient, added, deleted, changed_old)

		case "database":
			log.Debug("database")
			err = sendDatabaseHistory(notifierName, rec.Recipient, added, deleted, changed_old, changed_new)

		case "webhook":
			// TODO: implement webhook
			continue
		}

		if err != nil {
			log.Error("Failed to devliver notifier "+notifierName+" for recipient "+" recipient "+rec.Recipient, err)
			// TODO fail this upwards. maybe not return here, but let other notifiers run?
		}
	}
	log.Debug("save")
	// save updated calendar
	helpers.WriteCalFile(currentICS, historyFilename)
	return nil

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
		unsubscribeURL := conf.General.URL + "/view/" + url.QueryEscape(notifierName) + "/unsubscribe?mail=" + url.QueryEscape(recipient)
		m.SetHeader("List-Unsubscribe", unsubscribeURL)
		bodyunsubscribe := body + "\n\nUnsubscribe: " + unsubscribeURL
		m.SetBody("text/plain", string(bodyunsubscribe))
	}

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
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		feed.WriteRSS(file)

		return err
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return err
		}

		// parse existing feed
		feed, err = rss.ParseRSS(file)
		file.Close()
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

	// Write the updated RSS feed to the file
	file, err := os.Open(filename)
	if err != nil {
		return err
	}

	err = feed.WriteRSS(file)
	file.Close()

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
