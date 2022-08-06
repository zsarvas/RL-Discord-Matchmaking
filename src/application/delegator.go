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
	// case LEAVE_QUEUE:
	// 	cd.handleLeaveQueue(queue, player)
	// case REPORT_WIN:
	// 	cd.session.ChannelMessageSend(cd.discordUser.ChannelID, "Team wins.")
	// case QUEUE_STATUS:
	// 	// TODO(@Ritter_Gustave): refactor this later
	// 	presentationQueue := queue.DisplayQueue()
	// 	if presentationQueue == "" {
	// 		cd.session.ChannelMessageSend(cd.discordUser.ChannelID, "Queue is empty")
	// 		return
	// 	}

	// 	cd.session.ChannelMessageSend(cd.discordUser.ChannelID, queue.DisplayQueue())
	// case MATT:
	// 	cd.session.ChannelMessageSend(cd.discordUser.ChannelID, "Matt is a dingus.")
	default:
	}
}

func (d *Delegator) handleEnterQueue() {
	incomingId := d.DiscordUser.Author.String()
	prospectivePlayer := d.PlayerRepository.Get(incomingId)

	fmt.Printf("Incoming ID %s", incomingId)
	fmt.Printf("prospectivePlayer %v", prospectivePlayer)
	if d.queue.PlayerInQueue(prospectivePlayer) {
		formattedMessage := fmt.Sprintf("Player %s is already in the queue.", prospectivePlayer.DisplayName)
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, formattedMessage)

		return
	}

	d.queue.Enqueue(prospectivePlayer)
	formattedMessage := fmt.Sprintf("Player %s has entered the queue.", prospectivePlayer.DisplayName)
	d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, formattedMessage)
}

// 	// Queue should return if the addition of this player has popped the queue
// 	// and handle accordingly
// 	queuePopped := q.Enqueue(player)

// 	if queuePopped {
// 		// Dump the queue into a match
// 		matchHandler.DumpQueueIntoMatch(q)
// 		return
// 	}

// 	formattedMessage := fmt.Sprintf("Player %s has entered the queue.", player.DisplayName)
// 	cd.session.ChannelMessageSend(cd.discordUser.ChannelID, formattedMessage)
// }

// func (cd *command_delegator) handleLeaveQueue(q *queue.Queue, player player.Player) {
// 	playerSuccessfullyRemoved := q.LeaveQueue(player)

// 	if playerSuccessfullyRemoved {
// 		cd.session.ChannelMessageSend(
// 			cd.discordUser.ChannelID,
// 			fmt.Sprintf("Player %s has been removed from the queue.", player.DisplayName),
// 		)

// 		return
// 	}

// 	// If player is not in the queue, do nothing
// }
