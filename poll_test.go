package main

import (
	"fmt"
	tb "gopkg.in/tucnak/telebot.v2"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"
)

type TeleMockError struct {
}

func (TeleMockError) Error() string {
	return "test"
}

type TeleBotMock struct {
}

func (TeleBotMock) Send(to tb.Recipient, what interface{}, options ...interface{}) (*tb.Message, error) {
	return &tb.Message{ID:1}, nil
}
func (TeleBotMock) Handle(endpoint interface{}, handler interface{}) {
}
func (TeleBotMock) Edit(message tb.Editable, what interface{}, options ...interface{}) (*tb.Message, error) {
	return &tb.Message{ID:1}, nil
}
func (TeleBotMock) Respond(callback *tb.Callback, responseOptional ...*tb.CallbackResponse) error {
	return nil
}
func (TeleBotMock) Pin(message tb.Editable, options ...interface{}) error {
	return nil
}

func (TeleBotMock) ChatMemberOf(chat *tb.Chat, user *tb.User) (*tb.ChatMember, error) {
	return &tb.ChatMember{}, TeleMockError{}
}

func (TeleBotMock) Start() {
}

func (TeleBotMock) Stop() {
}

func (TeleBotMock) ChatByID(id string) (*tb.Chat, error) {
	return &tb.Chat{ID:1,FirstName:"x", LastName:"y", Username:"z"}, nil
}

func CreatePoll(users []string) *Poll {
	config := loadEnvConfiguration(true)
	bot, _ := NewBot(config.Telegram.Token);
	bot.Tbot = TeleBotMock{}
	return NewPoll("pic.jpg", randSeq(5), "Test poll", users, 10, time.Now().Add(time.Duration(1000000000000)), bot,"58547223");
}

func CreateVoter() *tb.User {
	return &tb.User{rand.Int(),
		randSeq(5),
		randSeq(5),
		randSeq(5),
		"ru"}
}

func TestPollCreate(t *testing.T) {
	fmt.Println("Testing creating poll")
	p := CreatePoll([]string{"kroshilin", "ivanov", "petrov"});
	if len(p.membersList) != 3 {
		t.Errorf("Count of members does not match provided!")
	}
}

func TestCheckIfUserMember(t *testing.T) {
	fmt.Println("Testing membership verifier")
	p := CreatePoll([]string{"1", "2", "3"});

	voter := CreateVoter();
	voter.ID= 1;

	if !p.checkIfUserIsMember(voter)  {
		t.Errorf("Membership verifying error!")
	}

	voter1 := CreateVoter();
	voter.ID= 6;

	if p.checkIfUserIsMember(voter1)  {
		t.Errorf("Membership verifying error!")
	}
}

func TestVoteFilter(t *testing.T) {
	fmt.Println("Testing vote filter")

	p := CreatePoll([]string{"1", "2"});

	voter := CreateVoter();
	voter.ID = 1;

	buttonId, response := p.filterVote(voter, btnYesId);
	if buttonId != btnYesId {
		t.Errorf("Failed to check that club member voted YES and correctly passed filter!")
		fmt.Println(buttonId, response)
	}

	buttonId3, _ := p.filterVote(voter, btnNoId);
	if buttonId3 != btnNoId {
		t.Errorf("Failed to check that club member voted NO and correctly passed filter!")
	}

	voter2 := CreateVoter();
	voter2.ID = 3;

	buttonId2, response2 := p.filterVote(voter2, btnYesId);

	if buttonId2 != pseudoBtnQueue {
		t.Errorf("Failed to check that NON member voted YES and queued before opening!")
		fmt.Println(buttonId2, response2)
	}

	buttonId3, response3 := p.filterVote(voter2, btnNoId);

	if buttonId3 != btnNoId {
		t.Errorf("Failed to check that NON member voted NO and was added to NO map!")
		fmt.Println(buttonId3, response3)
	}

	p.pollOpensForEveryoneAt = time.Now();

	buttonId4, response4 := p.filterVote(voter2, btnYesId);
	if buttonId4 != btnYesId {
		t.Errorf("Failed to check that NON member voted YES and was added to yes map after opening!")
		fmt.Println(buttonId4, response4)
	}
}

func TestPlayersRedistributionAfterOpen(t *testing.T) {
	fmt.Println("Testing players redistribution")

	p := CreatePoll([]string{"1", "2"});
	p.playersLimit = 4

	voter1 := CreateVoter();
	voter1.ID = 1;
	p.onVote(voter1, p.pollId + btnYesId);

	voter2 := CreateVoter();
	voter2.ID = 2;
	p.onVote(voter2, p.pollId + btnYesId);

	if len(p.results[btnYesId]) != 2 {
		t.Errorf("Fail to assert that there are two YES voters!")
	}

	voter3 := CreateVoter();
	voter3.ID = 3;
	p.onVote(voter3, p.pollId + btnYesId);

	voter4 := CreateVoter();
	voter4.ID = 4;
	p.onVote(voter4, p.pollId + btnYesId);

	if len(p.results[pseudoBtnQueue]) != 2 {
		t.Errorf("Fail to assert that there are two queued voters!")
	}

	if len(p.results[btnYesId]) != 2 {
		t.Errorf("Fail to assert that there are two YES voters!")
	}

	p.redistributeVotesOnOpenForEveryone()

	if len(p.results[btnYesId]) != 4 {
		t.Errorf("Fail to assert that voters were redistributed!")
	}
}

func setupFullPoll(members []string, fake bool) (*Poll, *Bot)  {
	os.Setenv("TZ", "Europe/Nicosia")
	rand.Seed(time.Now().UnixNano())
	config := loadEnvConfiguration(true)

	bot, _ := NewBot(config.Telegram.Token);
	if fake {
		bot.Tbot = TeleBotMock{}
	}
	picturer := picturer{&http.Client{}, config.PicturerApi}
	picture := picturer.GiveMePictureOf(config.PictureTags)

	t,_ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	opensIn := config.pollOpensForEveryoneBeforeEnd
	if !fake {
		opensIn += 10;
	}
	opensAt := t.Add(time.Second * time.Duration(opensIn))
	p := postPoll(t.Format("01/02 15:04") + "\nText description", picture, bot, config.PollRecipientId, members, config.playersLimit, opensAt)
	bot.addButtonsHandlers(p.buttons, p.onVote)
	if fake {
		time.AfterFunc(time.Duration(time.Second*time.Duration(config.pollOpensForEveryoneBeforeEnd+1)), bot.Tbot.Stop)
	}
	bot.Tbot.Start()
	return p, bot
}

func TestRealPoll(t *testing.T) {
	real, _ := strconv.Atoi(os.Getenv("REAL_POLL"))
	if real == 0 {
		return
	}
	config := loadEnvConfiguration(true)
	setupFullPoll(config.clubMembers, false)
}