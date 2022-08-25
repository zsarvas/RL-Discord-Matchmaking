package main

import (
	"errors"
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

var Token string
var sqlPath string
var playerRepository *interfaces.PlayerRepo
var matchRepository *interfaces.MatchRepo

func init() {
	flag.StringVar(&Token, "t", "", "Bot Token")
	flag.Parse()

	// Token Initialization
	err := godotenv.Load("dev.env")
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	Token = os.Getenv("TOKEN")

	if Token == "" {
		err := errors.New("no token found")
		log.Fatal(err)
	}

	sqlPath = os.Getenv("SQL_PATH")

	// Data Initialization
	playerRepoHandler := infrastructure.NewPlayerHandler(sqlPath)
	matchRepoHandler := infrastructure.NewMatchHandler()
	playerRepository = interfaces.NewPlayerRepo(playerRepoHandler)
	matchRepository = interfaces.NewMatchDataRepo(matchRepoHandler)

}

func main() {
	clientConnection, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Create application bot delegator
	// Register handler Function
	d := application.NewDelegator(playerRepository, matchRepository)
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
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)

	<-quit
	clientConnection.Close()
}
