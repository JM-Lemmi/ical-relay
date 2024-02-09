package main

import (
	"fmt"
	"reflect"
	"testing"
)

// prettyPrintConf - pretty prints the config struct
func prettyPrintConf(c Config) string {
	return fmt.Sprintf("%+v\n", c)
}

// TestParseConfig - tests that the config is parsed correctly
func TestParseConfig(t *testing.T) {
	// test that the config is parsed correctly
	conf, err := ParseConfig("./fixtures/testconfig.yaml")
	if err != nil {
		t.Errorf("Error parsing config: %s", err)
	}

	test_conf := Config{
		Version: 2,
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
