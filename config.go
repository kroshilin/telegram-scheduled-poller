package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"github.com/joho/godotenv"
)

type Config struct {
	Google   GoogleConfig
	Telegram TelegramConfig
	PictureTags    string
	PicturerApi    string
	PollRecipientId string
	clubMembers []string
	playersLimit int
	pollOpensForEveryoneBeforeEnd int
	CheckAndPostOnStart bool
}

type GoogleConfig struct {
	Key               []byte
	Email, CalendarId, HolidaysCalendarId string
}

type TelegramConfig struct {
	Token string
}

func loadEnvConfiguration(isTest bool) Config {
	log.Println("Loading configuration from .env")

	err := godotenv.Load()
	if isTest == true {
		err = godotenv.Overload(".env.test")
	}
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	email := os.Getenv("GOOGLE_SERVICE_ACCOUNT_EMAIL")
	key, err := (ioutil.ReadFile("private.key"))
	volleyCalendarId := os.Getenv("VOLLEY_CALENDAR_ID")
	holidaysCalendarId := os.Getenv("CYPRUS_HOLIDAYS_CALENDAR_ID")
	token := os.Getenv("TELEGRAM_TOKEN")
	tags := os.Getenv("TAGS_FOR_PICTURE")
	picturerApiUrl := os.Getenv("PICTURER_API_URL")
	clubMembers := strings.Split(os.Getenv("CLUB_MEMBERS"), ",")
	pollOpensForEveryoneBeforeEnd, _ := strconv.Atoi(os.Getenv("OPEN_POLL_FOR_EVERYONE_BEFORE_END"))
	playersLimit, _ := strconv.Atoi(os.Getenv("PLAYERS_LIMIT"))
	checkAndPostOnStart := os.Getenv("CHECK_AND_POST_ON_START") == "1"

	return Config{
		GoogleConfig{
			[]byte(key),
			email,
			volleyCalendarId,
			holidaysCalendarId,
		},
		TelegramConfig{token},
		tags,
		picturerApiUrl,
		os.Getenv("POLL_RECIPIENT_ID"),
		clubMembers,
		playersLimit,
		pollOpensForEveryoneBeforeEnd,
		checkAndPostOnStart,
	}
}
