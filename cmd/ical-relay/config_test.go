package main

import (
	"github.com/jm-lemmi/ical-relay/database"

	"fmt"
	"reflect"
	"testing"
)

// prettyPrintConf - pretty prints the config struct
func prettyPrintConf(c Config) string {
	return fmt.Sprintf("\n%+v\n", c)
}

// TestParseConfig - tests that the config is parsed correctly
func TestParseConfig(t *testing.T) {
	// test that the config is parsed correctly
	conf, err := ParseConfig("./fixtures/testconfig.yaml")
	if err != nil {
		t.Errorf("Error parsing config: %s", err)
	}

	test_conf := Config{
		Version: 3,
		Server: serverConfig{
			Addr:          ":80",
			FaviconPath:   "/static/media/favicon.svg",
			Imprint:       "https://your-imprint",
			LogLevel:      4,
			Name:          "Calendar",
			PrivacyPolicy: "http://your-data-privacy-policy",
			StoragePath:   "teststoragepath/",
			TemplatePath:  "/opt/ical-relay/templates/",
			URL:           "https://cal.julian-lemmerich.de",
			DB: dbConfig{
				Host:     "postgres",
				DbName:   "ical_relay",
				User:     "dbuser",
				Password: "password",
			},
			Mail: mailConfig{
				SMTPServer: "mailout.julian-lemmerich.de",
				SMTPPort:   25,
				Sender:     "calnotification@julian-lemmerich.de",
			},
			SuperTokens: []string{
				"rA4nhdhmr34lL6x6bLyGoJSHE9o9cA2BwjsMOeqV5SEzm61apcRRzWybtGVjLKiB",
			},
		},
	}

	if !reflect.DeepEqual(conf, test_conf) {
		t.Errorf("Config test failed, got %s -- should be %s", prettyPrintConf(conf), prettyPrintConf(test_conf))
	}
}

func TestExampleConfig(t *testing.T) {
	// test that the config is parsed correctly
	conf, err := ParseConfig("./config.yml.example")
	if err != nil {
		t.Errorf("Error parsing config: %s", err)
	}

	test_conf := Config{
		Version: 3,
		Server: serverConfig{
			Addr:          ":80",
			FaviconPath:   "/static/media/favicon.svg",
			Imprint:       "https://your-imprint",
			LogLevel:      4,
			Name:          "Calendar",
			PrivacyPolicy: "http://your-data-privacy-policy",
			StoragePath:   "./",
			TemplatePath:  "/opt/ical-relay/templates/",
			URL:           "https://cal.julian-lemmerich.de",
			DB: dbConfig{
				Host:     "postgres",
				DbName:   "ical_relay",
				User:     "dbuser",
				Password: "password",
			},
			Mail: mailConfig{
				SMTPServer: "mailout.julian-lemmerich.de",
				SMTPPort:   25,
				Sender:     "calnotification@julian-lemmerich.de",
			},
			SuperTokens: []string{
				"rA4nhdhmr34lL6x6bLyGoJSHE9o9cA2BwjsMOeqV5SEzm61apcRRzWybtGVjLKiB",
			},
		},
		Profiles: map[string]database.Profile{
			// 	"relay": profile{
			"relay": {
				Name:          "",
				Sources:       nil,
				Public:        true,
				ImmutablePast: true,
				Tokens: []database.Token{
					{
						Token: "eAn97Sa0BKHKk02O12lNsa1O5wXmqXAKrBYxRcTNsvZoU9tU4OVS6FH7EP4yFbEt",
					},
				},
				Rules: []database.Rule{
					{
						Filters: []map[string]string{
							{"regex": "testentry", "target": "summary", "type": "regex"},
							{"from": "2021-12-02T00:00:00Z", "type": "timeframe", "until": "2021-12-31T00:00:00Z"},
						},
						Operator: "",
						Action:   map[string]string{"type": "delete"},
						Expiry:   "",
					},
				},
			},
		},
		Notifiers: map[string]database.Notifier{
			"relay": {
				Name:     "",
				Source:   "http://localhost/relay",
				Interval: "15m",
				Recipients: []string{
					"jm.lemmerich@gmail.com",
				},
			},
		},
	}
	if !reflect.DeepEqual(conf, test_conf) {
		t.Errorf("Config test failed, got %s -- should be %s", prettyPrintConf(conf), prettyPrintConf(test_conf))
	}
}
