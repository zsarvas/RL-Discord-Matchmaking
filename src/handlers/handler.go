package handlers

import (
	"fmt"
	"strings"

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
			prospectivePlayer = strings.Split(prospectivePlayer, "#")[0]
			formattedMessage := fmt.Sprintf("Player %s is already in the queue.", prospectivePlayer)
			s.ChannelMessageSend(m.ChannelID, formattedMessage)
		} else {
			matchQueue.Enqueue(prospectivePlayer)
			prospectivePlayer = strings.Split(prospectivePlayer, "#")[0]
			formattedMessage := fmt.Sprintf("Player %s has entered the queue.", prospectivePlayer)
			s.ChannelMessageSend(m.ChannelID, formattedMessage)
		}

	case command.LEAVE_QUEUE:
		prospectivePlayer := m.Author.String()

		if matchQueue.PlayerInQueue(prospectivePlayer) {
			matchQueue.LeaveQueue(prospectivePlayer)
			prospectivePlayer = strings.Split(prospectivePlayer, "#")[0]
			formattedMessage := fmt.Sprintf("Player %s has been removed from the queue.", prospectivePlayer)
			s.ChannelMessageSend(m.ChannelID, formattedMessage)
		} else {
			prospectivePlayer = strings.Split(prospectivePlayer, "#")[0]
			formattedMessage := fmt.Sprintf("Player %s is not in the queue.", prospectivePlayer)
			s.ChannelMessageSend(m.ChannelID, formattedMessage)
		}

	case command.QUEUE_STATUS:
		currentQueue := matchQueue.DisplayQueue()
		s.ChannelMessageSend(m.ChannelID, currentQueue)

	case command.MATT:
		s.ChannelMessageSend(m.ChannelID, "Matt is a dingus.")

	case command.REPORT_WIN:
		s.ChannelMessageSend(m.ChannelID, "Team wins.")

	default:
		s.ChannelMessageSend(m.ChannelID, "Skippy is a dingus!")
	}
}
