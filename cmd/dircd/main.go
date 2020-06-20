package main

import (
	irc "github.com/btnmasher/dircd"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var log = logrus.New()

func init() {

	formatter := &prefixed.TextFormatter{
		ForceColors:      true,
		DisableSorting:   true,
		QuoteEmptyFields: true,
		FullTimestamp:    true,
	}

	log.SetFormatter(formatter)
	log.SetLevel(logrus.DebugLevel)

	irc.Warmup(log)
}

func main() {

	// Setup server and start
	server := irc.NewServer()
	server.SetHostname("irc.localhost.net")
	server.SetNetwork("dircd.net")

	//server.ListenAndServeTLS("server.pem", "server.key")
	server.ListenAndServe()
}
