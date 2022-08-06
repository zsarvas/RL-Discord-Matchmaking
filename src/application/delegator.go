package application

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/zsarvas/RL-Discord-Matchmaking/domain"
)

type Delegator struct {
	Session          *discordgo.Session
	DiscordUser      *discordgo.MessageCreate
	queue            *domain.Queue
	PlayerRepository domain.PlayerRepository
	command          string
}

func NewDelegator(playerRepo domain.PlayerRepository) *Delegator {
	// could possible move this queue out of the 'constructor'
	newQueue := domain.NewQueue(3)

	cd := &Delegator{
		PlayerRepository: playerRepo,
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
	default:
		return
	}
}

// Will check the DB or memory-implementation for a player
// If no player exists, makes a new one and returns it
func (d Delegator) fetchPlayer() domain.Player {
	incomingId := d.DiscordUser.Author.String()
	prospectivePlayer := d.PlayerRepository.Get(incomingId)

	return prospectivePlayer
}

func (d *Delegator) handleEnterQueue() {
	prospectivePlayer := d.fetchPlayer()

	if d.queue.PlayerInQueue(prospectivePlayer) {
		formattedMessage := fmt.Sprintf("Player %s is already in the queue.", prospectivePlayer.DisplayName)
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, formattedMessage)

		return
	}

	d.queue.Enqueue(prospectivePlayer)
	formattedMessage := fmt.Sprintf("Player %s has entered the queue.", prospectivePlayer.DisplayName)
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
			fmt.Sprintf("Player %s has been removed from the queue.", prospectivePlayer.DisplayName),
		)
	}
}

func (d Delegator) handleDisplayQueue() {
	presentationqueue := d.queue.DisplayQueue()

	if presentationqueue == "" {
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "queue is empty")
		return
	}

	d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, presentationqueue)
}
