package handlers

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

const ENTER_QUEUE string = "!q"
const LEAVE_QUEUE string = "!leave"
const REPORT_WIN string = "!report win"
const QUEUE_STATUS string = "!status"
const MATT string = "Matt"

var queue []string

func validatePlayerInQueue(player string) bool {
	for _, val := range queue {
		if val == player {
			return true
		}
	}

	return false
}

func convertQueueToReadableString() string {
	displayQueue := []string{}
	for _, val := range queue {
		santizedName := strings.Split(val, "#")[0]
		displayQueue = append(displayQueue, santizedName)
	}

	currentQueue := strings.Join(displayQueue, ", ")

	return currentQueue
}

func MessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Early return if the message is from the bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "What is Skippy?" {
		s.ChannelMessageSend(m.ChannelID, "A dingus!")
	}

	contentMessage := m.Content

	switch contentMessage {
	case ENTER_QUEUE:
		messagingPlayer := m.Author.String()

		if !validatePlayerInQueue(messagingPlayer) {
			queue = append(queue, messagingPlayer)
			s.ChannelMessageSend(m.ChannelID, "You have entered the queue")
		} else {
			formattedMessage := fmt.Sprintf("Player %s already in queue", messagingPlayer)
			s.ChannelMessageSend(m.ChannelID, formattedMessage)
		}

	case LEAVE_QUEUE:
		ks

	case QUEUE_STATUS:
		currentQueue := convertQueueToReadableString()
		s.ChannelMessageSend(m.ChannelID, currentQueue)

	case MATT:
		s.ChannelMessageSend(m.ChannelID, "Matt is a dingus")

	case REPORT_WIN:
		s.ChannelMessageSend(m.ChannelID, "Team wins.")

	default:
		s.ChannelMessageSend(m.ChannelID, "Skippy is a dingus!")
	}
}
