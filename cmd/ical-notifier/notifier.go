package main

import (
	"fmt"
	"net/url"
	"time"

	"gopkg.in/gomail.v2"

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

		// TODO: add different notification types

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

		for _, recipient := range notifier.Recipients {
			m := gomail.NewMessage()
			m.SetHeader("From", conf.General.Mail.Sender)
			m.SetHeader("To", recipient)
			m.SetHeader("Subject", "Calendar Notification for "+notifierName)

			unsubscribeURL := conf.General.URL + "/notifier/" + url.QueryEscape(notifierName) + "/unsubscribe?mail=" + url.QueryEscape(recipient)
			m.SetHeader("List-Unsubscribe", unsubscribeURL)
			bodyunsubscribe := body + "\n\nUnsubscribe: " + unsubscribeURL
			m.SetBody("text/plain", string(bodyunsubscribe))

			d := gomail.Dialer{Host: conf.General.Mail.SMTPServer, Port: conf.General.Mail.SMTPPort}
			if conf.General.Mail.SMTPUser != "" && conf.General.Mail.SMTPPass != "" {
				d = gomail.Dialer{Host: conf.General.Mail.SMTPServer, Port: conf.General.Mail.SMTPPort, Username: conf.General.Mail.SMTPUser, Password: conf.General.Mail.SMTPPass}
			}
			log.Info("Sending Mail Notification to " + recipient)

			err := d.DialAndSend(m)
			if err != nil {
				requestLogger.Errorln("error sending mail: " + err.Error())
				return err
			}
		}

		// save updated calendar
		helpers.WriteCalFile(historyICS, historyFilename)
		return nil
	}
}

// runs a heartbeat loop with specified sleep duration
func NotifierTiming(id string, interval time.Duration) {
	if interval == 0 {
		// failsave for 0s interval, to make machine still responsive
		interval = 1 * time.Second
	}
	log.Debug("Notifier " + id + ", Interval: " + interval.String())
	// endless loop
	for {
		time.Sleep(interval)
		notifyChanges(id)
	}
}

// starts a heartbeat notifier in a sub-routine
func NotifierStartup() {
	log.Info("Starting Notifiers")
	for id, n := range conf.getNotifiers() {
		interval, err := time.ParseDuration(n.Interval)
		if err != nil {
			log.Error("Failed to parse duration for notifier " + id + ": " + n.Interval)
		}
		go NotifierTiming(id, interval)
	}
}

func RunNotifier(id string) error {
	if !conf.notifierExists(id) {
		return fmt.Errorf("notifier not found")
	}
	return notifyChanges(id)
}
