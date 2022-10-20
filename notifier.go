package main

import (
	"os"
	"time"
	"fmt"

	"gopkg.in/gomail.v2"

	ics "github.com/arran4/golang-ical"
	log "github.com/sirupsen/logrus"
)

func notifyChanges(id string, n *notifier) error {
	requestLogger := log.WithFields(log.Fields{"notifier": id})
	requestLogger.Infoln("Running Notifier!")

	// check if file exists, if not download for the first time
	if _, err := os.Stat("/app/notifystore/" + id + ".ics"); os.IsNotExist(err) {
		log.Info("File does not exist, downloading for the first time")
		calendar, err := readCalURL(n.Source)
		if err != nil {
			requestLogger.Errorln(err)
			return err
		}
		// save file
		writeCalFile(calendar, "/app/notifystore/"+id+".ics")
	}

	// read files
	file1, _ := os.Open("/app/notifystore/" + id + ".ics")
	calendar1, err := ics.ParseCalendar(file1)
	if err != nil {
		requestLogger.Errorln(err)
		return err
	}

	calendar2, err := readCalURL(n.Source)
	if err != nil {
		requestLogger.Errorln(err)
		return err
	}

	added, deleted, changed := compare(calendar1, calendar2)

	if len(added) + len(deleted) + len(changed) == 0 {
		log.Info("No changes detected.")
		return nil
	} else {
		log.Debug("Changes detected: " + fmt.Sprint(len(added)) + " added, " + fmt.Sprint(len(deleted)) + " deleted, " + fmt.Sprint(len(changed)) + " changed")

		var body string

		if len(added) > 0 {
			for _, event := range added {
				body += "Added:\n\n" + event.Serialize() + "\n\n"
			}
		}
		if len(deleted) > 0 {
			for _, event := range deleted {
				body += "Deleted:\n\n" + event.Serialize() + "\n\n"
			}
		}
		if len(changed) > 0 {
			for _, event := range changed {
				body += "Changed (displaying new version):\n\n" + event.Serialize() + "\n\n"
			}
		}

		for _, recipient := range n.Recipients {
			m := gomail.NewMessage()
			m.SetHeader("From", conf.Server.Mail.Sender)
			m.SetHeader("To", recipient)
			m.SetHeader("Subject", "Calendar Notification for "+id)

			unsubscribeURL := conf.Server.URL + "/api/notifier/" + id + "/removerecipient?email=" + recipient
			m.SetHeader("List-Unsubscribe", unsubscribeURL)
			bodyunsubscribe := body + "\n\nUnsubscribe: " + unsubscribeURL
			m.SetBody("text/plain", string(bodyunsubscribe))

			d := gomail.Dialer{Host: conf.Server.Mail.SMTPServer, Port: conf.Server.Mail.SMTPPort}
			if conf.Server.Mail.SMTPUser != "" && conf.Server.Mail.SMTPPass != "" {
				d = gomail.Dialer{Host: conf.Server.Mail.SMTPServer, Port: conf.Server.Mail.SMTPPort, Username: conf.Server.Mail.SMTPUser, Password: conf.Server.Mail.SMTPPass}
			}
			log.Info("Sending Mail Notification to " + recipient)

			err := d.DialAndSend(m)
			if err != nil {
				requestLogger.Errorln(err)
				return err
			}
		}

		// save updated calendar
		writeCalFile(calendar2, "/app/notifystore/"+id+".ics")
		return nil
	}
	return fmt.Errorf("Impossible return location. If you get this error, something is wrong.")
}

// runs a heartbeat loop with specified sleep duration
func NotifierTiming(id string, n *notifier) {
	interval, _ := time.ParseDuration(n.Interval)
	if interval == 0 {
		// failsave for 0s interval, to make machine still responsive
		interval = 1 * time.Second
	}
	log.Debug("Notifier " + id + ", Interval: " + interval.String())
	// endless loop
	for {
		time.Sleep(interval)
		notifyChanges(id, n)
	}
}

// starts a heartbeat notifier in a sub-routine
func NotifierStartup() {
	log.Info("Starting Notifiers")
	for id, n := range conf.Notifiers {
		go NotifierTiming(id, &n)
	}
}

func RunNotifier(id string) error {
	n, ok := conf.Notifiers[id]
	if !ok {
		return fmt.Errorf("Notifier not found")
	}
	return notifyChanges(id, &n)
}
