package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/zsarvas/RL-Discord-Matchmaking/application"
	"github.com/zsarvas/RL-Discord-Matchmaking/infrastructure"
	"github.com/zsarvas/RL-Discord-Matchmaking/interfaces"
)

var (
	Token string
)

var playerRepository *interfaces.PlayerRepo

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()
	playerRepoHandler := infrastructure.NewPlayerHandler()
	playerRepository = interfaces.NewPlayerRepo(playerRepoHandler)
}

func main() {

	err := godotenv.Load("dev.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	Token := os.Getenv("TOKEN")

	if Token == "" {
		return
	}

	clientConnection, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	d := application.NewDelegator(playerRepository)
	// Registers handler Function
	clientConnection.AddHandler(d.InitiateDelegator)

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
