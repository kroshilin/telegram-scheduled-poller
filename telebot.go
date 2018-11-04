package main

import (
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"time"
)

type Bot struct {
	EventText    string
	EventPicture string
	Token        string
	Tbot         *tb.Bot
	Polls        map[int]SendablePoll
}

var cyprusHolyday = "Сегодня очередной кипрский праздник - ничего не будет 🤬"

func NewBot(Token string) (*Bot, error) {
	Tbot, err := tb.NewBot(tb.Settings{
		Token:  Token,
		Poller: &tb.LongPoller{Timeout: 10 * time.Second},
	})

	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	b := Bot{Tbot: Tbot, Polls:make(map[int]SendablePoll)}
	//b.Tbot.Start()

	return  &b, nil
}

func (b Bot) addButtonsHandlers(buttons map[string]tb.InlineButton, callback func(string, string)) {
	// Command: /start <PAYLOAD>
	//b.Tbot.Handle("/start", func(m *tb.Message) {
	//	if !m.Private() {
	//		return
	//	}
	//
	//	photo := &tb.Photo{
	//		Caption: b.EventText,
	//		File:    tb.FromURL(b.EventPicture),
	//	}
	//	b.Tbot.Send(m.Sender, photo, &tb.ReplyMarkup{
	//		InlineKeyboard: inlineKeys,
	//	})
	//})

	for _, v := range buttons {
		func (cl tb.InlineButton) {
			b.Tbot.Handle(&cl, func(c *tb.Callback) {
				b.Tbot.Respond(c, &tb.CallbackResponse{CallbackID: c.ID, Text: "Я тебя запомнил, " + c.Sender.Username})
				b.Tbot.Send(Recipient{fmt.Sprint(c.Message.Chat.ID)}, c.Sender.FirstName + " " + c.Sender.LastName + " проголосовал(a)")
				callback(c.Sender.Username, cl.Unique)
				b.UpdatePoll(c.Message)
			})
		}(v)
	}
}

func (b Bot) PostMessage(message string, recipient Recipient) {
	b.Tbot.Send(recipient, message)
}

func (b Bot) PostPoll(poll SendablePoll, recipient Recipient) {
	message, _ := b.Tbot.Send(recipient, poll.GetText(), &tb.SendOptions{
		ReplyMarkup:&tb.ReplyMarkup{InlineKeyboard: poll.GetLayout()},
		ParseMode: tb.ParseMode(tb.ModeMarkdown),
	})
	b.Polls[message.ID] = poll
	b.Tbot.Pin(message)
}

func (b Bot) UpdatePoll(message *tb.Message) {
	b.Tbot.Edit(message, b.Polls[message.ID].GetText(), &tb.SendOptions{
		ReplyMarkup:&tb.ReplyMarkup{InlineKeyboard:  b.Polls[message.ID].GetLayout()},
		ParseMode: tb.ParseMode(tb.ModeMarkdown),
	})
}