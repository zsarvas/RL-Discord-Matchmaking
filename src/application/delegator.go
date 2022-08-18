package application

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/zsarvas/RL-Discord-Matchmaking/domain"
)

type Delegator struct {
	Session          *discordgo.Session
	DiscordUser      *discordgo.MessageCreate
	queue            *domain.Queue
	PlayerRepository domain.PlayerRepository
	MatchRepository  MatchRepository
	command          string
}

func NewDelegator(playerRepo domain.PlayerRepository, matchRepo MatchRepository) *Delegator {
	// could possible move this queue out of the 'constructor'
	newQueue := domain.NewQueue(4)

	cd := &Delegator{
		PlayerRepository: playerRepo,
		MatchRepository:  matchRepo,
		queue:            newQueue,
	}

	return cd
}

func (d *Delegator) InitiateDelegator(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.Session = s
	d.DiscordUser = m
	d.command = m.Content

	d.HandleIncomingCommand()
}

func (d *Delegator) HandleIncomingCommand() {
	switch d.command {
	case ENTER_QUEUE:
		d.handleEnterQueue()
	case LEAVE_QUEUE:
		d.handleLeaveQueue()
	case QUEUE_STATUS:
		d.handleDisplayQueue()
	case REPORT_WIN:
		// Not Implemented Fully
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "Team wins.")
	case MATT:
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "Matt is a dingus.")
	case DISPLAY_MATCHES:
		// Should refactor, put this logic in appropriate layer
		activeMatches := d.MatchRepository.GetMatches()

		if len(activeMatches) == 0 {
			d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "No Active Matches")
			return
		}

		message := []string{}

		for k, v := range activeMatches {
			stringifiedTeamOne := []string{"["}
			stringifiedTeamTwo := []string{"["}

			for _, player := range v.TeamOne {
				stringifiedTeamOne = append(stringifiedTeamOne, player.DisplayName, ",")
			}
			for _, player := range v.TeamTwo {
				stringifiedTeamTwo = append(stringifiedTeamTwo, player.DisplayName, ",")
			}
			stringifiedTeamOne = append(stringifiedTeamOne, "]")
			stringifiedTeamTwo = append(stringifiedTeamTwo, "]")

			matchInformation := fmt.Sprintf(
				"Match id %s: between %s and %s \n", k.String(),
				strings.Join(stringifiedTeamOne, " "),
				strings.Join(stringifiedTeamTwo, " "),
			)

			message = append(message, matchInformation)
		}

		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, strings.Join(message, "\n"))
	default:
		return
	}
}

// Will check the DB or memory-implementation for a player
// If no player exists, makes a new one and returns it
func (d Delegator) fetchPlayer() domain.Player {
	incomingId := d.DiscordUser.Author.String()

	//fmt.Printf("fixed id: '%v'", fixedId)
	prospectivePlayer := d.PlayerRepository.Get(incomingId)

	return prospectivePlayer
}

func (d *Delegator) handleEnterQueue() {
	prospectivePlayer := d.fetchPlayer()

	if d.queue.PlayerInQueue(prospectivePlayer) {
		formattedMessage := fmt.Sprintf("%s is already in the queue.", prospectivePlayer.DisplayName)
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, formattedMessage)

		return
	}

	d.queue.Enqueue(prospectivePlayer)

	queueIsPopping := d.handleQueuePop()

	if queueIsPopping {
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "Queue POPPED!")
		match := Match{
			TeamOne: []domain.Player{d.queue.Dequeue(), d.queue.Dequeue()},
			TeamTwo: []domain.Player{d.queue.Dequeue(), d.queue.Dequeue()},
		}

		d.MatchRepository.Add(match)
		return
	}

	formattedMessage := fmt.Sprintf("%s has entered the queue.", prospectivePlayer.DisplayName)
	d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, formattedMessage)
}

func (d *Delegator) handleLeaveQueue() {
	incomingId := d.DiscordUser.Author.String()
	prospectivePlayer := d.PlayerRepository.Get(incomingId)

	if !d.queue.PlayerInQueue(prospectivePlayer) {
		// Player isn't in queue, exit
		return
	}

	playerSuccessfullyRemoved := d.queue.LeaveQueue(prospectivePlayer)

	if playerSuccessfullyRemoved {
		d.Session.ChannelMessageSend(
			d.DiscordUser.ChannelID,
			fmt.Sprintf("%s has been removed from the queue.", prospectivePlayer.DisplayName),
		)
	}
}

func (d Delegator) handleDisplayQueue() {
	presentationqueue := d.queue.DisplayQueue()

	if presentationqueue == "" {
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "Queue is empty")
		return
	}

	d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, presentationqueue)
}

func (d *Delegator) handleQueuePop() bool {
	queueLength := d.queue.GetQueueLength()
	popLength := d.queue.GetPopLength()

	return queueLength == popLength
}
