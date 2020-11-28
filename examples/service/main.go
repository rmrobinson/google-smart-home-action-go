package main

import (
	"context"
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	action "github.com/rmrobinson/google-smart-home-action-go"
	"go.uber.org/zap"
	"golang.org/x/crypto/acme/autocert"
	"google.golang.org/api/homegraph/v1"
	"google.golang.org/api/option"
)

func main() {
	var (
		auth0Domain     = flag.String("auth0-domain", "", "The domain that Auth0 users will be coming from")
		letsEncryptHost = flag.String("letsencrypt-host", "", "The host name that LetsEncrypt will generate the cert for")
		agentUserID     = flag.String("agent-user-id", "", "The HomeGraph account user ID to synchronize state with")
		credsFile       = flag.String("creds-file", "", "The Google Service Account key file path")
	)
	flag.Parse()

	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	// Setup our authentication validator
	auth := &auth0Authenticator{
		logger: logger,
		domain: *auth0Domain,
		client: &http.Client{},
		tokens: map[string]string{},
	}

	// Setup the 'mock' service
	es := &echoService{
		logger: logger,
		lights: map[string]lightbulb{
			"123": {
				"123",
				"test light 1",
				false,
				40,
				struct {
					hue        float64
					saturation float64
					value      float64
				}{
					100,
					100,
					10,
				},
			},
			"456": {
				"456",
				"test light 2",
				false,
				40,
				struct {
					hue        float64
					saturation float64
					value      float64
				}{
					100,
					100,
					10,
				},
			},
		},
		receiver: receiver{
			"789",
			"test receiver",
			false,
			20,
			false,
			"input_1",
		},
		agentID: *agentUserID,
	}

	// Setup Google Assistant info
	ctx := context.Background()
	hgService, err := homegraph.NewService(ctx, option.WithCredentialsFile(*credsFile))
	if err != nil {
		logger.Fatal("err initializing homegraph",
			zap.Error(err),
		)
	}

	svc := action.NewService(logger, auth, es, hgService)
	es.service = svc

	// Register callback from Google
	http.HandleFunc(action.GoogleFulfillmentPath, svc.GoogleFulfillmentHandler)

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGUSR1)
	go func() {
		for {
			<-c
			logger.Debug("toggling light 1 via signal")
			es.toggleLight1()
		}
	}()

	// Setup LetsEncrypt
	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(*letsEncryptHost), //Your domain here
		Cache:      autocert.DirCache("certs"),               //Folder for storing certificates
	}

	server := &http.Server{
		Addr: ":https",
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
	}

	go http.ListenAndServe(":http", certManager.HTTPHandler(nil))

	logger.Info("listening")

	log.Fatal(server.ListenAndServeTLS("", ""))
}
