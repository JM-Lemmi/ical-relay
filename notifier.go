package main

import (
	"os"
	"time"

	"gopkg.in/gomail.v2"

	ics "github.com/arran4/golang-ical"
	log "github.com/sirupsen/logrus"
)

func notifyChanges(id string, n *notifier) error {
	requestLogger := log.WithFields(log.Fields{"notifier": id})
	requestLogger.Infoln("New Notifier run!")

	// check if file exists, if not download for the first time
	if _, err := os.Stat("/app/notifystore/" + id + ".ics"); os.IsNotExist(err) {
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
	calendar1, _ := ics.ParseCalendar(file1)

	calendar2, _ := readCalURL(n.Source)

	added, deleted, changed := compare(calendar1, calendar2)

	var body string

	if len(added) > 0 {
		for _, event := range added {
			body += "Added: " + event.GetProperty("Summary").Value + ", " + event.GetProperty("DStart").Value + "\n"
		}
	}
	if len(deleted) > 0 {
		for _, event := range deleted {
			body += "Deleted: " + event.GetProperty("Summary").Value + ", " + event.GetProperty("DStart").Value + "\n"
		}
	}
	if len(changed) > 0 {
		for _, event := range changed {
			body += "Changed: " + event.GetProperty("Summary").Value + ", " + event.GetProperty("DStart").Value + "\n"
		}
	}

	for _, recipient := range n.Recipients {
		m := gomail.NewMessage()
		m.SetHeader("From", n.Sender)
		m.SetHeader("To", recipient)
		m.SetHeader("Subject", "Calendar Notification for "+id)
		m.SetBody("text/plain", string(body))

		d := gomail.Dialer{Host: n.SMTPServer, Port: n.SMTPPort}
		if n.SMTPUser != "" && n.SMTPPass != "" {
			d = gomail.Dialer{Host: n.SMTPServer, Port: n.SMTPPort, Username: n.SMTPUser, Password: n.SMTPPass}
		}
		log.Debug("Mail Notification sent to " + recipient)

		return d.DialAndSend(m)
	}

	return nil
}

// runs a heartbeat loop with specified sleep duration
func NotifierTiming(id string, n *notifier) {
	interval, _ := time.ParseDuration(n.Interval)
	for {
		time.Sleep(interval)
		notifyChanges(id, n)
	}
}

// starts a heartbeat notifier in a sub-routine
func NotifierStartup(conf *Config) {
	for id, n := range conf.Notifiers {
		go NotifierTiming(id, &n)
	}
}
