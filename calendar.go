package main

import (
	"log"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/calendar/v3"
)

type EventsChecker struct {
	calendarService *calendar.Service
}

type EmptyEventError struct {
}

func (e EmptyEventError) Error() string {
	return "No event for this date"
}

//check if there is any event for date
func (e EventsChecker) getEventForDate(calendarId string, dateTime time.Time) (*calendar.Event, error) {
	beginOfDay := time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), 0, 0, 0, 0, dateTime.Location())
	endOfDay := time.Date(dateTime.Year(), dateTime.Month(), dateTime.Day(), 23, 59, 0, 0, dateTime.Location())
	formattedStart := beginOfDay.Format(time.RFC3339)
	formattedEnd := endOfDay.Format(time.RFC3339)
	cal, err := e.calendarService.Events.List(calendarId).
		TimeMin(formattedStart).
		TimeMax(formattedEnd).
		MaxResults(1).
		ShowDeleted(false).
		SingleEvents(true).
		Do()
	if err != nil {
		log.Println("Calendar id:" + calendarId)
		log.Fatalf("Unable to get calendar: %v", err)
	}

	if len(cal.Items) > 0 {
		return cal.Items[0], nil
	} else {
		return nil, EmptyEventError{}
	}
}

func initCalendarService(email string, key []byte) (*calendar.Service, error) {
	conf := &jwt.Config{
		Email:      email,
		PrivateKey: key,
		Scopes: []string{
			"https://www.googleapis.com/auth/calendar",
		},
		TokenURL: google.JWTTokenURL,
	}
	client := conf.Client(oauth2.NoContext)
	svc, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to create calendar service: %v", err)
	}

	return svc, err
}
