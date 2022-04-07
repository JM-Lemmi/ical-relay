package main

import (
	"net/http"
	"os"

	rice "github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var version = "1.0.0-rc2"

var configPath = "config.yml"
var conf Config

var router *mux.Router

var templateBox *rice.Box

func main() {
	if len(os.Args) >= 2 {
		configPath = os.Args[1]
	}

	log.Info("Welcome to ical-relay, version " + version)

	// load config
	var err error
	conf, err = ParseConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	log.SetLevel(conf.Server.LogLevel)

	// find template box
	log.Infoln("Finding box templates")
	templateBox, err = rice.FindBox("templates")
	if err != nil {
		log.Fatal(err)
	}

	// setup router
	router = mux.NewRouter()
	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/profiles/{profile}", profileHandler).Name("profile")
	router.HandleFunc("/profiles/{profile}/view", profileViewHandler)

	// listen and serve
	address := conf.Server.Addr
	log.Fatal(http.ListenAndServe(address, router))
}
