package lib

import (
	"fmt"
	"strconv"

	"github.com/segmentio/analytics-go"
)

type AnalyticsCredentials struct {
	WriteKey string
}

type IAnalyticsManager interface {
	IdentifyUser(userEmail string, properties map[string]interface{}) error
	Track(userId int32, eventName string, properties map[string]interface{}) error
}

func NewAnalyticsManager(WriteKey string) *AnalyticsManager {
	client := analytics.New(WriteKey)
	return &AnalyticsManager{
		Client: client,
	}
}

type AnalyticsManager struct {
	Client analytics.Client
}

func (a *AnalyticsManager) IdentifyUser(userEmail string, properties map[string]interface{}) error {
	err := a.Client.Enqueue(analytics.Identify{
		UserId: userEmail,
		Traits: properties,
	})
	if err != nil {
		CaptureSentryException(fmt.Sprintf("Error occurred while identifying user: %v on segment", userEmail))
		return err
	}
	return nil
}

func (a *AnalyticsManager) Track(userId int32, eventName string, properties map[string]interface{}) error {
	err := a.Client.Enqueue(analytics.Track{
		Event:      eventName,
		UserId:     strconv.Itoa(int(userId)),
		Properties: properties,
	})
	if err != nil {
		CaptureSentryException(fmt.Sprintf("Error occurred while executing the event %v for user: %v on segment", eventName, userId))
		return err
	}
	return nil
}
