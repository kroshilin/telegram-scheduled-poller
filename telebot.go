package main

import (
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"log"
	"strconv"
	"time"
)

type TeleBot interface {
	Send(to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, error)
	Handle(endpoint interface{}, handler interface{})
	Edit(message tb.Editable, what interface{}, options ...interface{}) (*tb.Message, error)
	Respond(callback *tb.Callback, responseOptional ...*tb.CallbackResponse) error
	Pin(message tb.Editable, options ...interface{}) error
	Start()
	Stop()
	ChatByID(id string) (*tb.Chat, error)
    ChatMemberOf(chat *tb.Chat, user *tb.User) (*tb.ChatMember, error)
}

type Bot struct {
	EventText    string
	EventPicture string
	Token        string
	Tbot         TeleBot
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

	return  &b, nil
}

func (b Bot) addButtonsHandlers(buttons map[string]tb.InlineButton, callback func(*tb.User, string) string) {
	for _, v := range buttons {
		func (cl tb.InlineButton) {
			b.Tbot.Handle(&cl, func(c *tb.Callback) {
				response := callback(c.Sender, cl.Unique)
				b.Tbot.Respond(c, &tb.CallbackResponse{CallbackID: c.ID, Text: response})
				b.UpdatePoll(c.Message)
			})
		}(v)
	}
}

func (b Bot) PostMessage(message string, recipient Recipient, options *tb.SendOptions) {
	b.Tbot.Send(recipient, message, options)
}

func (b Bot) PostPoll(poll SendablePoll, recipient Recipient) *tb.Message {
	message, error := b.Tbot.Send(recipient, poll.GetText(), &tb.SendOptions{
		ReplyMarkup:&tb.ReplyMarkup{InlineKeyboard: poll.GetLayout()},
		ParseMode: tb.ParseMode(tb.ModeHTML),
	})

	if error != nil {
		fmt.Println(error)
	}

	b.Polls[message.ID] = poll
	b.Tbot.Pin(message)

	return message
}

func (b Bot) UpdatePoll(message *tb.Message) {
	updatedMessage, error := b.Tbot.Edit(message, b.Polls[message.ID].GetText(), &tb.SendOptions{
		ReplyMarkup:&tb.ReplyMarkup{InlineKeyboard:  b.Polls[message.ID].GetLayout()},
		ParseMode: tb.ParseMode(tb.ModeHTML),
	})

	 if error == nil {
		 b.Tbot.Pin(updatedMessage)
	 } else {
		 fmt.Print(error)
	 }
}

func (b Bot) ChatMemberOf(id string, recipientId string) (string, string) {
	userId, _ := strconv.Atoi(id)
	chatId, _ := strconv.Atoi(recipientId)
	chatMember, err := b.Tbot.ChatMemberOf(&tb.Chat{ID:int64(chatId)}, &tb.User{ID:userId})
	if err!=nil {
		fmt.Println("Error on requesting chat", err)
		return "unknown", "unknown"
	}

	return chatMember.User.Username, chatMember.User.FirstName + " " + chatMember.User.LastName
}