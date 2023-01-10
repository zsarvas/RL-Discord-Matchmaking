package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
	"github.com/zsarvas/RL-Discord-Matchmaking/application"
	"github.com/zsarvas/RL-Discord-Matchmaking/infrastructure"
	"github.com/zsarvas/RL-Discord-Matchmaking/interfaces"
)

var Token string
var tokenApi string
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

	tokenApi = os.Getenv("SUPABASE_CONNECTION_STRING")

	// Data Initialization
	playerRepoHandler := infrastructure.NewPlayerHandler(tokenApi)
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
	counter := 1

	if counter == 1 {

		embed := &discordgo.MessageEmbed{
			Author:      &discordgo.MessageEmbedAuthor{},
			Color:       0x00ff00, // Green,
			Description: "this is a stupid test message",
			Image: &discordgo.MessageEmbedImage{
				URL: "",
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "",
			},
			Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
			Title:     "This is a stupid test title",
			Footer: &discordgo.MessageEmbedFooter{
				Text:    "Created by Zach Sarvas and Ritter Gustave",
				IconURL: "https://media-exp1.licdn.com/dms/image/C560BAQF24YrdYxKgpw/company-logo_200_200/0/1535555980728?e=1669852800&v=beta&t=D18WBZeNWIGnBMbEGWzg94kpIoOmKgCMf8SrboMk9iw",
			},
		}

		d.Session.ChannelMessageSendEmbed("1011004892418166877", embed)
		counter++
	}
	// Wait for kill signal to terminate
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, syscall.SIGTERM)

	<-quit
	clientConnection.Close()
}
