package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/zsarvas/RL-Discord-Matchmaking/handlers"
)

var (
	Token string
)

func init() {

	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
}

func main() {
	clientConnection, err := discordgo.New("Bot" + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Registers handler Function
	clientConnection.AddHandler(handlers.MessageHandler)

	// Open websocket begin listening handle error
	err = clientConnection.Open()
	if err != nil {
		fmt.Println("error opening session,", err)
		return
	}

	fmt.Println("Bot is open and listening...")

	// Wait for kill signal to terminate
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)

	<-quit
	clientConnection.Close()
}
