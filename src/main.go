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
var tokenApi string
var playerRepository *interfaces.PlayerRepo
var playerRepo1v1 *interfaces.PlayerRepo
var playerRepo2v2 *interfaces.PlayerRepo
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

	tokenApi = os.Getenv("POSTGRES_CONNECTION_STRING")

	// Data Initialization - separate handlers for 1v1 and 2v2 tables
	playerRepoHandler1v1 := infrastructure.NewPlayerHandler(tokenApi, "rocketleague_1v1")
	playerRepoHandler2v2 := infrastructure.NewPlayerHandler(tokenApi, "rocketleague_2v2")
	matchRepoHandler := infrastructure.NewMatchHandler()
	playerRepo1v1 = interfaces.NewPlayerRepo(playerRepoHandler1v1)
	playerRepo2v2 = interfaces.NewPlayerRepo(playerRepoHandler2v2)
	matchRepository = interfaces.NewMatchDataRepo(matchRepoHandler)

	// Keep old variable for backward compatibility (will use 2v2)
	playerRepository = playerRepo2v2

}

func main() {
	clientConnection, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}

	// Create application bot delegator
	// Register handler Function
	d := application.NewDelegator(playerRepo1v1, playerRepo2v2, matchRepository)

	clientConnection.AddHandler(d.InitiateDelegator)

	// Open websocket begin listening handle error
	err = clientConnection.Open()
	if err != nil {
		fmt.Println("error opening session,", err)
		return
	}

	//clientConnection.UpdateGameStatus(0, "Rocket League 2")
	err = clientConnection.UpdateStatusComplex(discordgo.UpdateStatusData{
		Activities: []*discordgo.Activity{
			&discordgo.Activity{
				Name: "your commands.",
				Type: 2,
			},
		},
		Status: "online",
	})
	if err != nil {
		fmt.Println("error updating status,", err)
	}
	fmt.Println("Bot is open and listening...")

	// Wait for kill signal to terminate
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)

	<-quit
	clientConnection.Close()
}
