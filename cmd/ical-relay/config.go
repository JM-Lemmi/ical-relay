package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

// STRUCTS
// !! breaking changes need to keep the old version in legacyconfig.go !!

// Config represents configuration for the application
type Config struct {
	Version int          `yaml:"version"`
	Server  serverConfig `yaml:"server"`
}

type serverConfig struct {
	Addr            string     `yaml:"addr"`
	URL             string     `yaml:"url"`
	LogLevel        log.Level  `yaml:"loglevel"`
	StoragePath     string     `yaml:"storagepath"`
	LiteMode        bool       `yaml:"litemode,omitempty"`
	DisableFrontend bool       `yaml:"disable-frontend,omitempty"`
	TemplatePath    string     `yaml:"templatepath"`
	FaviconPath     string     `yaml:"faviconpath,omitempty"`
	Name            string     `yaml:"name,omitempty"`
	Imprint         string     `yaml:"imprintlink"`
	PrivacyPolicy   string     `yaml:"privacypolicylink"`
	DB              dbConfig   `yaml:"db,omitempty"`
	Mail            mailConfig `yaml:"mail,omitempty"`
	SuperTokens     []string   `yaml:"super-tokens,omitempty"`
}

type dbConfig struct {
	Host     string `yaml:"host"`
	DbName   string `yaml:"db-name"`
	User     string `yaml:"user"`
	Password string `yaml:"password,omitempty"`
}

type mailConfig struct {
	SMTPServer string `yaml:"smtp_server"`
	SMTPPort   int    `yaml:"smtp_port"`
	Sender     string `yaml:"sender"`
	SMTPUser   string `yaml:"smtp_user,omitempty"`
	SMTPPass   string `yaml:"smtp_pass,omitempty"`
}

// CONFIG MANAGEMENT FUNCTIONS

// ParseConfig reads config from path and returns a Config struct
func ParseConfig(path string) (Config, error) {
	var tmpConfig Config

	yamlFile, err := os.ReadFile(path)
	if err != nil {
		log.Fatalf("Error Reading Config: %v ", err)
		return tmpConfig, err
	}

	err = yaml.Unmarshal(yamlFile, &tmpConfig)
	if err != nil {
		log.Fatalf("Error Parsing Config: %v\nSince v2.0.0-beta.9 there are breaking changes to the config and data structure. Please update your config and data manually!", err) // TODO update version on v2.0.0 release
		return tmpConfig, err
	}

	// check if config has current version, if not upgrade it
	if tmpConfig.Version < 4 {
		log.Warn("Config is outdated but could be parsed! Since v2.0.0-beta.9 there are breaking changes to the config and data structure. Please update your config and data manually!") // TODO update version on v2.0.0 release
	}

	log.Trace("Read config, now setting defaults")
	// defaults
	if tmpConfig.Server.Addr == "" {
		tmpConfig.Server.Addr = ":8080"
	}
	if tmpConfig.Server.LogLevel == 0 {
		tmpConfig.Server.LogLevel = log.InfoLevel
	}
	if tmpConfig.Server.StoragePath == "" {
		tmpConfig.Server.StoragePath = filepath.Dir(path)
	}
	if !strings.HasSuffix(tmpConfig.Server.StoragePath, "/") {
		tmpConfig.Server.StoragePath += "/"
	}
	if tmpConfig.Server.TemplatePath == "" {
		tmpConfig.Server.TemplatePath = filepath.Dir("/opt/ical-relay/templates/")
	}
	if !strings.HasSuffix(tmpConfig.Server.TemplatePath, "/") {
		tmpConfig.Server.TemplatePath += "/"
	}
	if tmpConfig.Server.Name == "" {
		tmpConfig.Server.Name = "Calendar"
	}
	if tmpConfig.Server.FaviconPath == "" {
		tmpConfig.Server.FaviconPath = "/static/media/favicon.svg"
	}

	return tmpConfig, nil
}

func (c Config) saveConfig(path string) error {
	currentConfig, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	err = os.WriteFile(
		path[:strings.LastIndexByte(path, '.')]+time.Now().UTC().Format("2006-01-02_150405")+".bak.yml",
		currentConfig,
		0600)
	if err != nil {
		return err
	}

	d, err := yaml.Marshal(&c)
	if err != nil {
		return err
	}
	return os.WriteFile(path, d, 0600)
}
