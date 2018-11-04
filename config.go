package main

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/joho/godotenv"
)

const answerYes = "Я в игре"
const answerNo = "Я слишком стар для этого"
const answerMaybe = "Не уверен"
const answerPlus2 = "+2"
const answerPlus3 = "+3 o_O"

type Config struct {
	Google   GoogleConfig
	Telegram TelegramConfig
	Shutterstock    ShutterstockConfig
	PollRecipientId string
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

func loadEnvConfiguration() Config {
	log.Println("Loading configuration from .env")

	err := godotenv.Load()
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
	}
}
