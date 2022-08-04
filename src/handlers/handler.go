package handlers

import (
	"github.com/bwmarrin/discordgo"
	"github.com/zsarvas/RL-Discord-Matchmaking/command"
	"github.com/zsarvas/RL-Discord-Matchmaking/queue"
)

var matchQueue queue.Queue

func MessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Create new command delegator
	cmd_delegator := command.NewDelegator(s, m)

	// Early return if the message is from the bot
	if m.Author.ID == s.State.User.ID {
		return
	}

	// Delegate and handle command appropriately
	// Augment the matchQueue as needed
	cmd_delegator.HandleIncomingCommand(&matchQueue)
}
