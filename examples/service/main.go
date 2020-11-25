package main

import (
	"flag"
	"fmt"
	"net/http"

	action "github.com/rmrobinson/google-smart-home-action-go"
	"go.uber.org/zap"
)

func main() {
	var (
		port = flag.Int("port", 443, "The port to listen on")
	)
	flag.Parse()

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Setup Google Assistant info
	svc := action.NewService(logger)

	// Register callback from Google
	http.HandleFunc(action.GoogleFulfillmentPath, svc.GoogleFulfillmentHandler)

	err := http.ListenAndServe(fmt.Sprintf(":%d", *port), nil)
	if err != nil {
		logger.Fatal("listenAndServe error",
			zap.Error(err),
		)
	}

}
