package main

import (
	"github.com/jasonlvhit/gocron"
	tb "gopkg.in/tucnak/telebot.v2"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"
)

type Recipient struct {
	ChatId string
}
func (r Recipient) Recipient() string {
	return r.ChatId
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func main() {
	os.Setenv("TZ", "Europe/Nicosia")
	rand.Seed(time.Now().UnixNano())
	config := loadEnvConfiguration(false)

	bot, _ := NewBot(config.Telegram.Token);
	picturer := picturer{config.Shutterstock.Login, config.Shutterstock.Password, &http.Client{}}
	calendarService, _ := initCalendarService(config.Google.Email, config.Google.Key)
	checker := EventsChecker{calendarService}

	checkOffsetForWeekend, _ := time.ParseDuration("12h10m10s") // means that event will be checked for next calendar day
	checkOffsetForWeekday, _ := time.ParseDuration("0h0m0s")
	gocron.Every(1).Friday().At("15:00").Do(checkAndPostPoll, picturer, checker, bot, config, checkOffsetForWeekend)
	gocron.Every(1).Saturday().At("15:00").Do(checkAndPostPoll, picturer, checker, bot, config, checkOffsetForWeekend)
	gocron.Every(1).Monday().At("09:00").Do(checkAndPostPoll, picturer, checker, bot, config, checkOffsetForWeekday)
	gocron.Every(1).Wednesday().At("09:00").Do(checkAndPostPoll, picturer, checker, bot, config, checkOffsetForWeekday)
	if config.CheckAndPostOnStart {
		checkAndPostPoll(picturer, checker, bot, config, checkOffsetForWeekday)
	}
	gocron.Start()
	bot.Tbot.Start()
}

func checkAndPostPoll(picturer picturer, checker EventsChecker, bot *Bot, config Config, checkOffset time.Duration) {
	picture := picturer.GiveMePictureOf(config.Shutterstock.Tags)
	volleyEvent, _ := checker.getEventForDate(config.Google.CalendarId, time.Now().Add(checkOffset))
	holiday, _ := checker.getEventForDate(config.Google.HolidaysCalendarId, time.Now().Add(checkOffset))

	membersList := []string{}
	if holiday != nil {
		date, _ := time.Parse(time.RFC3339, holiday.Start.Date)

		if date.Weekday().String() == "Sunday" || date.Weekday().String() == "Saturday" {
			holiday = nil
		} else {
			volleyEvent = nil
		}
	} else {
		if volleyEvent != nil {
			if  !strings.Contains(strings.ToLower(volleyEvent.Description), "пляж") && !strings.Contains(strings.ToLower(volleyEvent.Summary), "пляж") {
				membersList = config.clubMembers
			}
		}
	}

	if volleyEvent != nil {
		t, _ := time.Parse(time.RFC3339, volleyEvent.Start.DateTime)

		opensAt := t
		if len(membersList) > 0 {
			opensAt = t.Add(time.Second * time.Duration(config.pollOpensForEveryoneBeforeEnd) * -1)
		}

		postPoll(t.Format("01/02 15:04") + "\n" + volleyEvent.Description, picture, bot, config.PollRecipientId, membersList, config.playersLimit, opensAt)
	}

	if holiday != nil {
		// post sad message about cyprus holiday
		bot.PostMessage(cyprusHolyday, Recipient{config.PollRecipientId}, &tb.SendOptions{
			ParseMode: tb.ParseMode(tb.ModeHTML),
		})
	}
}

func postPoll(text string, picture string, bot *Bot, recipient string, membersList []string,playersLimit int, pollOpensForEveryoneAt time.Time) *Poll {
	poll := NewPoll(picture, randSeq(5), text, membersList, playersLimit, pollOpensForEveryoneAt, bot, recipient)
	bot.addButtonsHandlers(poll.buttons, poll.onVote)
	poll.originalMessage = bot.PostPoll(poll, Recipient{recipient})
	return poll
}