package lib

import (
	"context"
	"log"
	"time"

	"github.com/getsentry/sentry-go"
)

func InitiateSentry(config *Config) {
	// Sentry is dependent on env and needs to be initaited even before config.json is fetched from AWS.
	DSN := config.DSN
	ENV := config.ENV
	TraceSampleRate := config.TraceSampleRate
	if len(DSN) == 0 {
		log.Fatal("ERROR: DSN not found")
	}
	if len(ENV) == 0 {
		log.Fatal("ERROR: ENV not found")
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:           DSN,
		Environment:   ENV,
		EnableTracing: true,
		// Set TracesSampleRate to 1.0 to capture 100%
		// of transactions for performance monitoring.
		// We recommend adjusting this value in production,
		TracesSampleRate: TraceSampleRate,
	})
	if err != nil {
		log.Fatalf("sentry.Init: %s", err)
	} else {
		log.Println("INFO: Sentry initiated successfully")
	}
	// Flush buffered events before the program terminates.
	defer sentry.Flush(2 * time.Second)
}

func CreateSpan(context *context.Context, title string) (func(), *sentry.Span) {
	span := sentry.StartSpan(*context, title)
	return func() {
		span.Finish()
	}, span
}

func CreateTransaction(title string) (func(), *sentry.Span) {
	ctx := context.Background()

	hub := sentry.GetHubFromContext(ctx)
	if hub == nil {
		hub = sentry.CurrentHub().Clone()
		ctx = sentry.SetHubOnContext(ctx, hub)
	}

	span := sentry.StartTransaction(ctx, title)
	return func() {
		span.Finish()
	}, span
}

func AddSentryTag(req *Request, key string, value string) {
	if hub := sentry.GetHubFromContext(req.SentryContext); hub != nil {
		hub.Scope().SetTag(key, value)
	}
}
