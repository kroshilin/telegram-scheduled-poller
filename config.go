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
	Shutterstock    ShutterstockConfig
	PollRecipientId string
	clubMembers []string
	playersLimit int
	pollOpensForEveryoneBeforeEnd int
}

type GoogleConfig struct {
	Key               []byte
	Email, CalendarId, HolidaysCalendarId string
}

type TelegramConfig struct {
	Token string
}

type ShutterstockConfig struct {
	Login, Password, Tags string
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
	shtLogin := os.Getenv("SHUTTERSTOCK_LOGIN")
	shtPassword := os.Getenv("SHUTTERSTOCK_PASSWORD")
	shtTags := os.Getenv("TAGS_FOR_SHUTTERSTOCK")
	clubMembers := strings.Split(os.Getenv("CLUB_MEMBERS"), ",")
	pollOpensForEveryoneBeforeEnd, _ := strconv.Atoi(os.Getenv("OPEN_POLL_FOR_EVERYONE_BEFORE_END"))
	playersLimit, _ := strconv.Atoi(os.Getenv("PLAYERS_LIMIT"))

	return Config{
		GoogleConfig{
			[]byte(key),
			email,
			volleyCalendarId,
			holidaysCalendarId,
		},
		TelegramConfig{token},
		ShutterstockConfig{shtLogin, shtPassword, shtTags},
		os.Getenv("POLL_RECIPIENT_ID"),
		clubMembers,
		playersLimit,
		pollOpensForEveryoneBeforeEnd,
	}
}
