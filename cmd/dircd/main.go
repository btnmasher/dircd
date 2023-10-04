/*
   Copyright (c) 2023, btnmasher
   All rights reserved.
   Use of this source code is governed by a BSD-style
   license that can be found in the LICENSE file.
*/

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sourcegraph/conc"

	irc "github.com/btnmasher/dircd"

	"github.com/sirupsen/logrus"
)

func main() {
	mainContext, shutdown := context.WithCancel(context.Background())
	defer shutdown()

	wg := conc.NewWaitGroup()
	defer wg.Wait()

	shutdownTimeout := 30 * time.Second
	logger := logrus.New()
	//logger.SetReportCaller(true)

	// Setup server and start
	server, cfgErr := irc.NewServer(
		irc.WithHostname("irc.localhost.net"),
		irc.WithNetwork("dircd.net"),
		irc.WithLogger(logger),
		irc.WithLogLevel(logrus.DebugLevel),
		irc.WithDefaultLogFormatter(),
		irc.WithGracefulShutdown(mainContext, shutdownTimeout),
	)
	if cfgErr != nil {
		logger.Fatal(cfgErr)
	}
	wg.Go(func() {
		//server.ListenAndServeTLS("server.pem", "server.key")
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, irc.ErrServerClosed) {
			logger.Fatal(fmt.Errorf("failed to start server: %w", err))
		}
	})

	log := logger.WithField("component", "main")
	killSignals := make(chan os.Signal, 1)
	signal.Notify(killSignals, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-killSignals
		log.Infof("initializing server shutdown, received signal: %s", sig)
		shutdown()
		sig = <-killSignals
		log.Fatalf("forcefully shutting down server, received signal: %s", sig)
	}()
}
