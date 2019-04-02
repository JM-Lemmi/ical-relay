package main

import (
	"net/http"
	"os"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

var configPath = "config.toml"
var conf Config

func main() {
	// load config
	var err error
	conf, err = ParseConfig(configPath)
	if err != nil {
		log.Fatal(err)
	}

	log.SetLevel(log.DebugLevel)

	// setup router
	router := mux.NewRouter()
	router.HandleFunc("/", IndexHandler)
	router.HandleFunc("/profiles/{profile}", ProfileHandler)

	// listen and serve
	address := os.Args[1]
	log.Fatal(http.ListenAndServe(address, router))
}
