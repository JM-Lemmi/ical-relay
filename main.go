package main

import (
	"net/http"

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

	log.SetLevel(conf.Server.LogLevel)

	// setup router
	router := mux.NewRouter()
	router.HandleFunc("/", indexHandler)
	router.HandleFunc("/profiles/{profile}", profileHandler)

	// listen and serve
	address := conf.Server.Addr
	log.Fatal(http.ListenAndServe(address, router))
}
