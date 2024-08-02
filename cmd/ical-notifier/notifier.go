package main

import (
	"fmt"
	"net/url"
	"os"

	"gopkg.in/gomail.v2"

	ics "github.com/arran4/golang-ical"
	"github.com/gopherlibs/feedhub/feedhub"
	"github.com/jm-lemmi/ical-relay/compare"
	"github.com/jm-lemmi/ical-relay/helpers"
	log "github.com/sirupsen/logrus"
)

func notifyChanges(notifierName string, notifier notifier) error {
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
		return err
	}

	added, deleted, changed := compare.Compare(currentICS, historyICS)

	if len(added)+len(deleted)+len(changed) == 0 {
		log.Info("No changes detected.")
		return nil
	} else {
		log.Debug("Changes detected: " + fmt.Sprint(len(added)) + " added, " + fmt.Sprint(len(deleted)) + " deleted, " + fmt.Sprint(len(changed)) + " changed")

		// iterate over all recipients by type
		for rectype, recipients := range notifier.Recipients {
			switch rectype {
			case "mail":
				sendNotifyMails(notifierName, recipients, added, deleted, changed)

			case "rss":
				err := sendRSSFeed(notifierName, added, deleted, changed)
				if err != nil {
					log.Error("Failed to send RSS feed for notifier", notifierName, err)
				}

			case "webhook":
				// TODO: implement webhook
				continue
			}
		}

		// save updated calendar
		helpers.WriteCalFile(historyICS, historyFilename)
		return nil
	}
}

func sendNotifyMails(notifierName string, recipients []string, added []ics.VEvent, deleted []ics.VEvent, changed []ics.VEvent) {
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

	for _, recipient := range recipients {
		m := gomail.NewMessage()
		m.SetHeader("From", conf.General.Mail.Sender)
		m.SetHeader("To", recipient)
		m.SetHeader("Subject", "Calendar Notification for "+notifierName)

		if !conf.General.LiteMode {
			unsubscribeURL := conf.General.URL + "/notifier/" + url.QueryEscape(notifierName) + "/unsubscribe?mail=" + url.QueryEscape(recipient)
			m.SetHeader("List-Unsubscribe", unsubscribeURL)
			bodyunsubscribe := body + "\n\nUnsubscribe: " + unsubscribeURL
			m.SetBody("text/plain", string(bodyunsubscribe))
		}

		d := gomail.Dialer{Host: conf.General.Mail.SMTPServer, Port: conf.General.Mail.SMTPPort}
		if conf.General.Mail.SMTPUser != "" && conf.General.Mail.SMTPPass != "" {
			d = gomail.Dialer{Host: conf.General.Mail.SMTPServer, Port: conf.General.Mail.SMTPPort, Username: conf.General.Mail.SMTPUser, Password: conf.General.Mail.SMTPPass}
		}
		log.Info("Sending Mail Notification to " + recipient)

		err := d.DialAndSend(m)
		if err != nil {
			log.Errorln("error sending mail: " + err.Error())
		}
	}
}

func sendRSSFeed(notifierName string, added []ics.VEvent, deleted []ics.VEvent, changed []ics.VEvent) error {
	filename := conf.General.StoragePath + "rssstore/" + notifierName + ".rss"

	feed, err := helpers.LoadRSSFeed(filename)
	if err != nil {
		log.Error("Failed to load RSS feed for notifier", notifierName, err)
		return err
	}

	// add new items
	for _, event := range added {
		feed.Add(&feedhub.Item{
			Title:       "Added " + event.GetProperty(ics.ComponentPropertySummary).Value,
			Description: helpers.PrettyPrint(event),
		})
	}
	for _, event := range deleted {
		feed.Add(&feedhub.Item{
			Title:       "Deleted " + event.GetProperty(ics.ComponentPropertySummary).Value,
			Description: helpers.PrettyPrint(event),
		})
	}
	for _, event := range changed {
		feed.Add(&feedhub.Item{
			Title:       "Changed " + event.GetProperty(ics.ComponentPropertySummary).Value,
			Description: helpers.PrettyPrint(event),
		})
	}

	// Write the updated RSS feed to the file
	fhandle, err := os.Open(filename)
	if err != nil {
		return err
	}

	err = feed.WriteRss(fhandle)
	return err
}
