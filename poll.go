package main

import (
	"fmt"
	"html"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
	tb "gopkg.in/tucnak/telebot.v2"
)

type SendablePoll interface {
	GetText() string
	GetLayout() [][]tb.InlineButton
}

type Poll struct {
	eventPicture           string
	eventText              string
	pollId                 string
	results                map[string]map[int]Vote
	buttons                map[string]tb.InlineButton
	buttonsLayout          [][]tb.InlineButton
	membersList            []Member
	pollOpensForEveryoneAt time.Time
	bot                    *Bot
	recipient              string
	playersLimit           int
	originalMessage        *tb.Message
}

type Member struct {
	Id       string
	Username string
	Name     string
}

type Vote struct {
	voter        *tb.User
	vote         bool
	time         time.Time
	isClubMember bool
	username     string
}

type VoteResult struct {
	Count int
	Votes []Vote
}

const btnYesId = "yes"
const btnNoId = "no"
const btnMaybeId = "maybe"
const pseudoBtnQueue = "queue"

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

func (p Poll) GetText() string {
	membersList := []string{}
	for _, member := range p.membersList {
		membersList = append(membersList, getMentionText(member))
	}

	membersListText := "\n Ждем первых 8 человек, остальные занимают очередь. \n-----------"
	return "<a href='" + p.eventPicture + "'>\u200b</a>" + p.eventText + "\n " + membersListText + "\n" + p.pollResultsTemplate(p.mapPollResults())
}

func (p Poll) GetLayout() [][]tb.InlineButton {
	return p.buttonsLayout
}

func NewPoll(picture string, pollId string, text string, membersList []string, playersLimit int, pollOpensForEveryoneAt time.Time, bot *Bot, recipient string) *Poll {
	poll := Poll{eventPicture: picture, pollId: pollId}
	btns, layout := poll.createPollButtonsAndLayout()
	poll.buttonsLayout = layout
	poll.buttons = btns
	poll.eventText = text
	poll.results = make(map[string]map[int]Vote)
	poll.results[btnYesId] = make(map[int]Vote)
	poll.results[btnNoId] = make(map[int]Vote)
	poll.results[btnMaybeId] = make(map[int]Vote)
	poll.results[pseudoBtnQueue] = make(map[int]Vote)
	poll.pollOpensForEveryoneAt = pollOpensForEveryoneAt
	poll.bot = bot
	poll.recipient = recipient
	poll.playersLimit = playersLimit
	pollPointer := &poll
	poll.membersList = []Member{}
	if len(membersList) > 0 {
		for _, memberId := range membersList {
			username, name := poll.bot.ChatMemberOf(memberId, recipient)
			poll.membersList = append(poll.membersList, Member{memberId, username, name})
		}
		time.AfterFunc(pollOpensForEveryoneAt.Sub(time.Now()), func() { pollPointer.redistributeVotesOnOpenForEveryone() })
	}
	return pollPointer
}

func (p Poll) redistributeVotesOnOpenForEveryone() {
	fmt.Println("Redistributing votes")

	results := p.mapPollResults()
	var addedPlayers []string

	if results[btnYesId].Count < p.playersLimit {
		if results[pseudoBtnQueue].Count > 0 {
			playersToAdd := p.playersLimit - results[btnYesId].Count
			for _, vote := range results[pseudoBtnQueue].Votes {
				if playersToAdd == 0 {
					continue
				}
				addedPlayers = append(addedPlayers, getMentionText(Member{
					strconv.Itoa(vote.voter.ID),
					vote.voter.Username,
					vote.voter.FirstName + " " + vote.voter.LastName,
				}))
				p.results[btnYesId][vote.voter.ID] = vote
				delete(p.results[pseudoBtnQueue], vote.voter.ID)
				playersToAdd--
			}
		}
	}

	if results[btnYesId].Count < p.playersLimit {
		p.bot.PostMessage("Ну же, кто-нибудь еще!", Recipient{p.recipient}, &tb.SendOptions{
			ParseMode: tb.ParseMode(tb.ModeHTML),
		})
	} else {
		p.bot.PostMessage("Состав набран! Следите за обновлениями - есть шанс, что кто-то откажется в последний момент.",
			Recipient{p.recipient}, &tb.SendOptions{
				ParseMode: tb.ParseMode(tb.ModeHTML),
			})
	}

	if p.originalMessage != nil {
		p.bot.UpdatePoll(p.originalMessage)
	}
}

func (p Poll) mapPollResults() map[string]*VoteResult {
	resultMap := map[string]*VoteResult{"yes": &VoteResult{0, []Vote{}},
		"no":    &VoteResult{0, []Vote{}},
		"maybe": &VoteResult{0, []Vote{}},
		"queue": &VoteResult{0, []Vote{}},
	}
	for btnId, v := range p.results {
		for _, userVote := range v {
			if userVote.vote == true {
				userVoteToSave := Vote{
					userVote.voter,
					userVote.vote,
					userVote.time,
					p.checkIfUserIsMember(userVote.voter),
					getMentionText(Member{
						strconv.Itoa(userVote.voter.ID),
						userVote.voter.Username,
						userVote.voter.FirstName + " " + userVote.voter.LastName,
					}),
				}

				switch btnId {
				case btnYesId:
					resultMap[btnYesId].Count += 1
					resultMap[btnYesId].Votes = append(resultMap[btnYesId].Votes, userVoteToSave)
				case pseudoBtnQueue:
					resultMap[pseudoBtnQueue].Count += 1
					userVoteToSave.username += " (" + userVoteToSave.time.Format("15:04:05") + ")"
					resultMap[pseudoBtnQueue].Votes = append(resultMap[pseudoBtnQueue].Votes, userVoteToSave)
				case btnNoId:
					resultMap[btnNoId].Count += 1
					resultMap[btnNoId].Votes = append(resultMap[btnNoId].Votes, userVoteToSave)
				case btnMaybeId:
					resultMap[btnMaybeId].Count += 1
					resultMap[btnMaybeId].Votes = append(resultMap[btnMaybeId].Votes, userVoteToSave)
				}
			}
		}
	}

	for _, resultPart := range resultMap {
		sort.Slice(resultPart.Votes, func(i, j int) bool {
			return resultPart.Votes[i].time.Before(resultPart.Votes[j].time)
		})
	}
	return resultMap
}

func (p Poll) pollResultsTemplate(resultMap map[string]*VoteResult) string {
	resultsTemplate := ""
	var yesNames []string
	for _, voteResultUser := range resultMap[btnYesId].Votes {
		yesNames = append(yesNames, voteResultUser.username)
	}
	var noNames []string
	for _, voteResultUser := range resultMap[btnNoId].Votes {
		noNames = append(noNames, voteResultUser.username)
	}
	var queueNames []string
	for _, voteResultUser := range resultMap[pseudoBtnQueue].Votes {
		queueNames = append(queueNames, voteResultUser.username)
	}
	var maybeNames []string
	for _, voteResultUser := range resultMap[btnMaybeId].Votes {
		maybeNames = append(maybeNames, voteResultUser.username)
	}

	resultsTemplate += "<b>Придут</b> " + strconv.Itoa(resultMap["yes"].Count) + " | " + strings.Join(yesNames, ", ") + "\n" +
		"<b>В очереди</b> " + strconv.Itoa(resultMap["queue"].Count) + " | " + strings.Join(queueNames, ", ") + "\n" +
		"<b>Сомневаются</b> " + strconv.Itoa(resultMap["maybe"].Count) + " | " + strings.Join(maybeNames, ", ") + "\n" +
		"<b>Не придут</b> " + strconv.Itoa(resultMap["no"].Count) + " | " + strings.Join(noNames, ", ") + " "

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
		buttonsMap[btnMaybeId] = tb.InlineButton{Unique: p.pollId + btnNoId, Text: selectRandomOption(noIWontOptions)},
		tb.InlineButton{Unique: p.pollId + btnYesId, Text: selectRandomOption(yesIWillOptions)},
		tb.InlineButton{Unique: p.pollId + btnMaybeId, Text: selectRandomOption(maybeOptions)}

	layout := [][]tb.InlineButton{
		[]tb.InlineButton{buttonsMap[btnYesId]},
		[]tb.InlineButton{buttonsMap[btnMaybeId]},
		[]tb.InlineButton{buttonsMap[btnNoId]},
	}

	return buttonsMap, layout
}

func (p Poll) onVote(voter *tb.User, buttonId string) string {
	originalButtonId, response := strings.Replace(buttonId, p.pollId, "", 1), "Я тебя запомнил"
	if _, ok := p.results[originalButtonId][voter.ID]; ok {
		return "Да понял я, понял"
	}

	originalButtonId, response = p.filterVote(voter, originalButtonId)
	isClubMember := p.checkIfUserIsMember(voter)
	for btnId, _ := range p.results {
		if _, ok := p.results[btnId][voter.ID]; ok {
			delete(p.results[btnId], voter.ID)
		}
	}
	p.results[originalButtonId][voter.ID] = Vote{voter, true, time.Now(), isClubMember, voter.Username}

	return response
}

func (p Poll) filterVote(voter *tb.User, buttonId string) (string, string) {
	defaultResponse := "Я тебя запомнил, " + voter.Username
	results := p.mapPollResults()
	if results[btnYesId].Count >= p.playersLimit && time.Now().After(p.pollOpensForEveryoneAt) &&
		(buttonId == btnYesId) {
		return pseudoBtnQueue, "Состав набран. Но, возможно, кто-то откажется."
	}
	if p.checkIfUserIsMember(voter) {
		return buttonId, defaultResponse
	} else {
		if time.Now().After(p.pollOpensForEveryoneAt) {
			return buttonId, defaultResponse
		}
		if buttonId == btnYesId {
			return pseudoBtnQueue, "Я добавлю тебя в очередь, " + voter.Username
		}
	}

	return buttonId, defaultResponse
}

func (p Poll) checkIfUserIsMember(voter *tb.User) bool {
	for _, e := range p.membersList {
		if e.Id == strconv.Itoa(voter.ID) {
			return true
		}
	}

	return false
}

func getMentionText(member Member) string {
	var username string
	if len(member.Username) > 0 {
		username = member.Username
	} else {
		username = member.Name
	}

	t := transform.Chain(norm.NFC)
	username, _, _ = transform.String(t, username)
	username = html.EscapeString(username)
	return "<a href='tg://user?id=" + member.Id + "'>" + username + "</a>"
}
