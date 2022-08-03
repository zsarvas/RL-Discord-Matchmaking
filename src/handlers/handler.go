package handlers

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/zsarvas/RL-Discord-Matchmaking/command"
	"github.com/zsarvas/RL-Discord-Matchmaking/queue"
)

var matchQueue queue.Queue

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
	case command.ENTER_QUEUE:
		prospectivePlayer := m.Author.String()

		if matchQueue.PlayerInQueue(prospectivePlayer) {
			formattedMessage := fmt.Sprintf("Player %s already in queue", prospectivePlayer)
			s.ChannelMessageSend(m.ChannelID, formattedMessage)
		} else {
			matchQueue.Enqueue(prospectivePlayer)
			s.ChannelMessageSend(m.ChannelID, "You have entered the queue")
		}

	case command.QUEUE_STATUS:
		currentQueue := matchQueue.DisplayQueue()
		s.ChannelMessageSend(m.ChannelID, currentQueue)

	case command.MATT:
		s.ChannelMessageSend(m.ChannelID, "Matt is a dingus")

	case command.REPORT_WIN:
		s.ChannelMessageSend(m.ChannelID, "Team wins.")

	default:
		s.ChannelMessageSend(m.ChannelID, "Skippy is a dingus!")
	}
}
