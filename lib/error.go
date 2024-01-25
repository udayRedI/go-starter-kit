package lib

import (
	"errors"
	"fmt"
	"log"

	"github.com/getsentry/sentry-go"
)

// check helps avoid repetitive fatal error checking.
func CheckFatal(e error, msg string) {
	if e != nil {
		sentry.CaptureException(e)
		log.Fatal(msg)
	}
}

func CaptureSentryException(msg string) {
	err := errors.New(msg)
	log.Println(fmt.Sprintf("Error: %s", msg))
	sentry.CaptureException(err)
}
