package command

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/zsarvas/RL-Discord-Matchmaking/queue"
)

type command_delegator struct {
	session     *discordgo.Session
	discordUser *discordgo.MessageCreate
	playerId    string
	player      string
	command     string
}

func NewDelegator(s *discordgo.Session, r *discordgo.MessageCreate) *command_delegator {
	cd := &command_delegator{
		session:     s,
		discordUser: r,
		playerId:    r.Author.String(),
		player:      strings.Split(r.Author.String(), "#")[0],
		command:     r.Content,
	}

	return cd
}

func (cd *command_delegator) HandleIncomingCommand(queue *queue.Queue) {
	switch cd.command {
	case ENTER_QUEUE:
		cd.handleEnterQueue(queue)
	case LEAVE_QUEUE:
		cd.handleLeaveQueue(queue)
	case REPORT_WIN:
		cd.session.ChannelMessageSend(cd.discordUser.ChannelID, "Team wins.")
	case QUEUE_STATUS:
		// TODO(@Ritter_Gustave): refactor this later
		presentationQueue := queue.DisplayQueue()
		if presentationQueue == "" {
			cd.session.ChannelMessageSend(cd.discordUser.ChannelID, "Queue is empty")
			return
		}

		cd.session.ChannelMessageSend(cd.discordUser.ChannelID, queue.DisplayQueue())
	case MATT:
		cd.session.ChannelMessageSend(cd.discordUser.ChannelID, "Matt is a dingus.")
	default:
	}
}

func (cd *command_delegator) handleEnterQueue(q *queue.Queue) {
	if q.PlayerInQueue(cd.player) {
		formattedMessage := fmt.Sprintf("Player %s is already in the queue.", cd.player)
		cd.session.ChannelMessageSend(cd.discordUser.ChannelID, formattedMessage)

		return
	}

	q.Enqueue(cd.playerId)
	formattedMessage := fmt.Sprintf("Player %s has entered the queue.", cd.player)
	cd.session.ChannelMessageSend(cd.discordUser.ChannelID, formattedMessage)
}

func (cd *command_delegator) handleLeaveQueue(q *queue.Queue) {
	playerSuccessfullyRemoved := q.LeaveQueue(cd.playerId)

	if playerSuccessfullyRemoved {
		cd.session.ChannelMessageSend(
			cd.discordUser.ChannelID,
			fmt.Sprintf("Player %s has been removed from the queue.", cd.player),
		)

		return
	}

	// If player is not in the queue, do nothing
}
