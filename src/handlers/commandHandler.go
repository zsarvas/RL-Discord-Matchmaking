package handlers

import (
	"github.com/bwmarrin/discordgo"
)

func MessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Early return if the message is from the bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	if m.Content == "What is Skippy?" {
		s.ChannelMessageSend(m.ChannelID, "A dingus!")
	}
}
