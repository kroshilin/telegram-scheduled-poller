package main

import (
	"golang.org/x/text/transform"
	tb "gopkg.in/tucnak/telebot.v2"
	"html"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"golang.org/x/text/unicode/norm"
)

type SendablePoll interface {
	GetText() string
	GetLayout() [][]tb.InlineButton
}

type Poll struct {
	eventPicture string
	eventText string
	pollId string
	results map[string]map[int]Vote
	buttons map[string]tb.InlineButton
	buttonsLayout [][]tb.InlineButton
}

type Vote struct {
	voter *tb.User
	vote bool
}

const btnYesId = "yes"
const btnNoId = "no"
const btnMaybeId = "maybe"
const btnPlus2Id = "plus2"
const btnPlus3Id = "plus3"

var yesIWillOptions = []string{
	"Я в игре🤩",
	"Базара 0😝",
	"Будем жечь🤘",
	"Еще бы🤠",
	"100 процентов👻",
	"Я всю неделю этого ждал🤤",
	"Выезжаю!🚂",
	"Я в теме😎",
}

var noIWontOptions = []string{
	"Я не приду😭",
	"Я слишком стар для этого🧓",
	"Я в домике🙆‍",
	"Я хотел, но...🤦",
	"0 процентов🙅‍",
	"Играйте сами в свой волейбол⚽",
}

var maybeOptions = []string{
	"Я еще подумаю🤔",
	"Сомневаюсь🤥",
	"Буду знать попозже🤐",
	"Может быть😶",
}

var plusTwoOptions = []string{"+2😯"}
var plusThreeOptions = []string{"+3😲"}

func (p Poll) GetText() string {
	return "<a href='" + p.eventPicture + "'>\u200b</a>" + p.eventText + "\n" + p.pollResultsTemplate()
}

func (p Poll) GetLayout() [][]tb.InlineButton {
	return p.buttonsLayout
}

func NewPoll(picture string, pollId string, text string) *Poll {
	poll := Poll{eventPicture:picture, pollId: pollId}
	btns, layout := poll.createPollButtonsAndLayout()
	poll.buttonsLayout = layout
	poll.buttons = btns
	poll.eventText = text
	poll.results = make(map[string]map[int]Vote)
	return &poll
}

func (p Poll) pollResultsTemplate() string {
	type voteResult struct {
		Count int
		Usernames []string
	}
	resultMap := map[string]*voteResult{"yes" : &voteResult{0, []string{}},
		"no" : &voteResult{0, []string{}},
		"maybe" : &voteResult{0, []string{}},
	}
	for btnId, v := range p.results {
		for _, userVote := range v {
			if userVote.vote == true {
				var username string
				if len(userVote.voter.Username) > 0 {
					username = userVote.voter.Username
				} else {
					username = userVote.voter.FirstName + " " + userVote.voter.LastName
				}

				t := transform.Chain(norm.NFC)
				username, _, _ = transform.String(t, username)
				username = html.EscapeString(username)
				username = "<a href='tg://user?id=" + strconv.Itoa(userVote.voter.ID) + "'>" + username + "</a>"

				switch btnId {
				case btnYesId:
					resultMap["yes"].Count += 1
					resultMap["yes"].Usernames = append(resultMap["yes"].Usernames, username)
				case btnNoId:
					resultMap["no"].Count += 1
					resultMap["no"].Usernames = append(resultMap["no"].Usernames, username)
				case btnMaybeId:
					resultMap["maybe"].Count += 1
					resultMap["maybe"].Usernames = append(resultMap["maybe"].Usernames, username)
				case btnPlus2Id:
					resultMap["yes"].Count += 2
					resultMap["yes"].Usernames = append(resultMap["yes"].Usernames, username + " (2)")
				case btnPlus3Id:
					resultMap["yes"].Count += 3
					resultMap["yes"].Usernames = append(resultMap["yes"].Usernames, username + " (3)")
				}
			}
		}
	}

	resultsTemplate := "<b>Придут</b> " + strconv.Itoa(resultMap["yes"].Count) + " | " + strings.Join(resultMap["yes"].Usernames, ", ") + "\n" +
		"<b>Сомневаются</b> " + strconv.Itoa(resultMap["maybe"].Count) + " | " + strings.Join(resultMap["maybe"].Usernames, ", ") + "\n" +
		"<b>Не придут</b> " + strconv.Itoa(resultMap["no"].Count) + " | " + strings.Join(resultMap["no"].Usernames, ", ") + " "

	return resultsTemplate
}

func selectRandomOption(reasons []string) string {
	rand.Seed(time.Now().Unix())
	return reasons[rand.Intn(len(reasons))]
}

func (p Poll) createPollButtonsAndLayout() (map[string]tb.InlineButton, [][]tb.InlineButton) {
	buttonsMap := make(map[string]tb.InlineButton)
	buttonsMap[btnNoId],
	buttonsMap[btnYesId],
	buttonsMap[btnMaybeId],
	buttonsMap[btnPlus2Id],
	buttonsMap[btnPlus3Id] = tb.InlineButton{Unique: p.pollId + btnNoId, Text: selectRandomOption(noIWontOptions)},
		tb.InlineButton{Unique: p.pollId + btnYesId, Text: selectRandomOption(yesIWillOptions)},
		tb.InlineButton{Unique: p.pollId + btnMaybeId, Text: selectRandomOption(maybeOptions)},
		tb.InlineButton{Unique: p.pollId + btnPlus2Id, Text: selectRandomOption(plusTwoOptions)},
		tb.InlineButton{Unique: p.pollId + btnPlus3Id, Text: selectRandomOption(plusThreeOptions)}

	layout := [][]tb.InlineButton{
		[]tb.InlineButton{buttonsMap[btnYesId], buttonsMap[btnPlus2Id], buttonsMap[btnPlus3Id]},
		[]tb.InlineButton{buttonsMap[btnMaybeId], buttonsMap[btnNoId]},
	}

	return buttonsMap, layout
}

func (p Poll) onVote(voter *tb.User, buttonId string) {
	originalButtonId := strings.Replace(buttonId, p.pollId, "", 1)
	if p.results != nil {
		for i, _ := range p.results {
			p.results[i][voter.ID] = Vote{voter, false}
		}
	}
	if p.results[originalButtonId] == nil {
		p.results[originalButtonId] = make(map[int]Vote)
	}
	p.results[originalButtonId][voter.ID] = Vote{voter, true};
}