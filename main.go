package main

import (
	"log"
	"net/http"

	sentryhttp "github.com/getsentry/sentry-go/http"
	"github.com/udayRedI/go-starter-kit/apps/health"
	"github.com/udayRedI/go-starter-kit/lib"
)

func main() {
	config := lib.Config{}
	lib.GetSecrets(&config)

	apps := []lib.App{
		health.New(),
	}

	s := lib.NewService(&config, &apps)

	startPort := s.Init()

	log.Println("INFO: Server started on localhost" + startPort)
	handler := sentryhttp.New(sentryhttp.Options{}).Handle(s.Server.Handler)

	if config.ENV == "LOCALL" {
		startPort = "localhost" + startPort
	}

	log.Fatal(http.ListenAndServe(startPort, handler))
}
