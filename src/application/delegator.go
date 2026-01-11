package application

import (
	"errors"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"

	"encoding/json"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/google/uuid"
	"github.com/zsarvas/RL-Discord-Matchmaking/domain"
)

type Delegator struct {
	Session           *discordgo.Session
	DiscordUser       *discordgo.MessageCreate
	queue1v1          *domain.Queue
	queue2v2          *domain.Queue
	currentQueue      *domain.Queue
	currentQueueType  QueueType
	PlayerRepo1v1     domain.PlayerRepository
	PlayerRepo2v2     domain.PlayerRepository
	currentPlayerRepo domain.PlayerRepository
	MatchRepository   MatchRepository
	command           string
}

type AvatarDecorationData struct {
	Asset         string `json:"asset"`
	SkuID         string `json:"sku_id"`
	AvatarDetails []int  `json:"asset_details,omitempty"`
}

type DiscordUser struct {
	ID                   string                `json:"id"`
	Username             string                `json:"username"`
	Avatar               string                `json:"avatar"`
	Discriminator        string                `json:"discriminator"`
	PublicFlags          int                   `json:"public_flags"`
	PremiumType          int                   `json:"premium_type"`
	Flags                int                   `json:"flags"`
	Banner               *string               `json:"banner"`       // Pointer to handle null value
	AccentColor          *int                  `json:"accent_color"` // Pointer to handle null value
	GlobalName           string                `json:"global_name"`
	AvatarDecorationData *AvatarDecorationData `json:"avatar_decoration_data"` // Now using proper struct
	BannerColor          *string               `json:"banner_color"`           // Pointer to handle null value
}

const (
	PLAYER_ADD              int    = 0
	PLAYER_LEFT             int    = 1
	PLAYER_SHOW             int    = 3
	PLAYER_ALREADY_IN_QUEUE int    = 4
	PLAYER_ALREADY_IN_MATCH int    = 5
	PLAYER_NOT_IN_QUEUE     int    = 6
	DISPLAY_QUEUE           int    = 7
	DISPLAY_HELP_MENU       int    = 8
	LOGO_URL1               string = ""
	LOGO_URL2               string = ""
	ICON_URL                string = "logo.png"
	FOURMANSCHANNELID       string = "1011004892418166877"
	ONEVSONECHANNELID       string = "1455331680096354305"
	GUILDID                 string = "189628012604555265"
	ROLEID_2V2              string = "1028789594277302302"
	ROLEID_1V1              string = "1455374709582467084"
)

type QueueType string

const (
	QueueType1v1 QueueType = "1v1"
	QueueType2v2 QueueType = "2v2"
)

// TimerKey is a composite key for tracking timers per player per queue type
type TimerKey struct {
	PlayerID  int
	QueueType QueueType
}

var playerTimers = make(map[TimerKey]*time.Timer)

func NewDelegator(playerRepo1v1 domain.PlayerRepository, playerRepo2v2 domain.PlayerRepository, matchRepo MatchRepository) *Delegator {
	// Create separate queues for 1v1 (pops at 2) and 2v2 (pops at 4)
	queue1v1 := domain.NewQueue(2)
	queue2v2 := domain.NewQueue(4)

	cd := &Delegator{
		PlayerRepo1v1:     playerRepo1v1,
		PlayerRepo2v2:     playerRepo2v2,
		currentPlayerRepo: playerRepo2v2, // Default to 2v2
		MatchRepository:   matchRepo,
		queue1v1:          queue1v1,
		queue2v2:          queue2v2,
		currentQueue:      queue2v2, // Default to 2v2
		currentQueueType:  QueueType2v2,
	}

	return cd
}

func (d *Delegator) InitiateDelegator(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.Session = s
	d.DiscordUser = m
	d.command = strings.ToUpper(m.Content)

	// Route to appropriate queue based on channel
	if m.ChannelID == ONEVSONECHANNELID {
		d.currentQueue = d.queue1v1
		d.currentQueueType = QueueType1v1
		d.currentPlayerRepo = d.PlayerRepo1v1
	} else if m.ChannelID == FOURMANSCHANNELID {
		d.currentQueue = d.queue2v2
		d.currentQueueType = QueueType2v2
		d.currentPlayerRepo = d.PlayerRepo2v2
	} else {
		// Not a valid channel, ignore command
		d.command = ""
		return
	}

	if strings.Contains(d.command, REPORT_WIN) {
		d.command = REPORT_WIN
	}

	if strings.Contains(d.command, DISPLAY_LEADERBOARD) {
		d.command = DISPLAY_LEADERBOARD
	}

	d.HandleIncomingCommand()
}

func (d *Delegator) HandleIncomingCommand() {
	switch d.command {
	case ENTER_QUEUE:
		d.handleEnterQueue()
	case LEAVE_QUEUE:
		d.handleLeaveQueue()
	case QUEUE_STATUS:
		d.handleDisplayQueue()
	case REPORT_WIN:
		d.handleMatchOver()
	case CLEAR_QUEUE:
		d.handleClearQueue()
	case DISPLAY_HELP:
		d.handleDisplayHelp()
	case DISPLAY_LEADERBOARD:
		d.handleDisplayLeaderboard()
	case MATT:
		d.specialMattFunction()
	case DISPLAY_MATCHES:
		d.handleDisplayMatches()
	default:
		return
	}
}

// Will check the DB or memory-implementation for a player
// If no player exists, makes a new one and returns it
func (d Delegator) fetchPlayer() domain.Player {
	incomingDiscordId := d.DiscordUser.Author.ID
	strIncomingDiscordId, err := strconv.Atoi(incomingDiscordId)
	if err != nil {
		log.Fatal(err)
	}
	globalName, err := d.getGlobalName(d.DiscordUser.Author.ID, d.DiscordUser.Member.Nick)
	if err != nil {
		log.Fatal(err)
	}
	mention := d.DiscordUser.Author.Mention()
	prospectivePlayer := d.currentPlayerRepo.Get(globalName, strIncomingDiscordId)
	prospectivePlayer.MentionName = mention
	prospectivePlayer.Id = globalName

	return prospectivePlayer
}

func (d *Delegator) handleEnterQueue() {

	prospectivePlayer := d.fetchPlayer()
	d.startTimeoutTimer(prospectivePlayer)

	// Check if player is already in the current queue
	if d.currentQueue.PlayerInQueue(prospectivePlayer) {
		d.changeQueueMessage(PLAYER_ALREADY_IN_QUEUE, prospectivePlayer)
		return
	}

	// Check if player is in a match for THIS queue type (not the other queue)
	// Each queue type has its own MatchId, so they can be in a 1v1 match and still queue for 2v2
	if prospectivePlayer.MatchId != uuid.Nil {
		d.changeQueueMessage(PLAYER_ALREADY_IN_MATCH, prospectivePlayer)
		return
	}

	d.currentQueue.Enqueue(prospectivePlayer)

	queueIsPopping := d.handleQueuePop()

	if queueIsPopping {
		d.currentQueue.RandomizeQueue()

		var match Match
		if d.currentQueueType == QueueType1v1 {
			// 1v1: 2 players, one on each team
			match = Match{
				TeamOne: []domain.Player{d.currentQueue.Dequeue()},
				TeamTwo: []domain.Player{d.currentQueue.Dequeue()},
			}
		} else {
			// 2v2: 4 players, 2 on each team
			match = Match{
				TeamOne: []domain.Player{d.currentQueue.Dequeue(), d.currentQueue.Dequeue()},
				TeamTwo: []domain.Player{d.currentQueue.Dequeue(), d.currentQueue.Dequeue()},
			}
		}

		matchId := d.MatchRepository.Add(match)

		for _, player := range match.TeamOne {
			player.MatchId = matchId
			d.currentPlayerRepo.SetMatch(player)
		}

		for _, player := range match.TeamTwo {
			player.MatchId = matchId
			d.currentPlayerRepo.SetMatch(player)
		}
		//queue popped here
		d.handleLobbyReady()
		return
	}

	d.changeQueueMessage(PLAYER_ADD, prospectivePlayer)

	//go func() {
	//	select {
	//	case <-time.After(20 * time.Minute):
	//		if d.queue.PlayerInQueue(prospectivePlayer) {
	//			d.Session.ChannelMessageSend(FOURMANSCHANNELID, prospectivePlayer.MentionName+" has been timed out from the queue.")
	//			d.queue.LeaveQueue(prospectivePlayer)
	//			d.changeQueueMessage(PLAYER_LEFT, prospectivePlayer)
	//		}

	//	}
	//}()

}

func (d *Delegator) startTimeoutTimer(player domain.Player) {
	channelID := d.DiscordUser.ChannelID
	currentQueue := d.currentQueue
	currentQueueType := d.currentQueueType

	// Create a composite key for this player in this queue type
	timerKey := TimerKey{
		PlayerID:  player.DiscordId,
		QueueType: currentQueueType,
	}

	playerTimers[timerKey] = time.AfterFunc(20*time.Minute, func() {
		if currentQueue.PlayerInQueue(player) {
			d.Session.ChannelMessageSend(channelID, player.MentionName+" has been timed out from the queue.")
			currentQueue.LeaveQueue(player)
			d.changeQueueMessage(PLAYER_LEFT, player)
		}
		// Clean up the timer
		delete(playerTimers, timerKey)
	})
}

func (d *Delegator) stopTimeoutTimer(player domain.Player) {
	// Create a composite key for this player in the current queue type
	timerKey := TimerKey{
		PlayerID:  player.DiscordId,
		QueueType: d.currentQueueType,
	}

	if timer, ok := playerTimers[timerKey]; ok {
		timer.Stop()
		delete(playerTimers, timerKey)
	}
}

func (d *Delegator) handleLeaveQueue() {
	incomingDiscordId := d.DiscordUser.Author.ID
	strIncomingDiscordId, err := strconv.Atoi(incomingDiscordId)
	if err != nil {
		log.Fatal(err)
	}
	globalName, err := d.getGlobalName(d.DiscordUser.Author.ID, d.DiscordUser.Member.Nick)
	if err != nil {
		log.Fatal(err)
	}
	mention := d.DiscordUser.Author.Mention()
	prospectivePlayer := d.currentPlayerRepo.Get(globalName, strIncomingDiscordId)
	prospectivePlayer.MentionName = mention
	prospectivePlayer.Id = globalName

	if !d.currentQueue.PlayerInQueue(prospectivePlayer) {
		// Player isn't in queue, exit
		d.changeQueueMessage(PLAYER_NOT_IN_QUEUE, prospectivePlayer)
		return
	}

	playerSuccessfullyRemoved := d.currentQueue.LeaveQueue(prospectivePlayer)

	if playerSuccessfullyRemoved {
		d.changeQueueMessage(PLAYER_LEFT, prospectivePlayer)
	}

	d.stopTimeoutTimer(prospectivePlayer)
}

func (d Delegator) handleDisplayQueue() {
	incomingDiscordId := d.DiscordUser.Author.ID
	presentationqueue := d.currentQueue.DisplayQueue()
	strIncomingDiscordId, err := strconv.Atoi(incomingDiscordId)
	if err != nil {
		log.Fatal(err)
	}
	globalName, err := d.getGlobalName(d.DiscordUser.Author.ID, d.DiscordUser.Member.Nick)
	if err != nil {
		log.Fatal(err)
	}
	callingPlayer := d.currentPlayerRepo.Get(globalName, strIncomingDiscordId)

	if presentationqueue == "" {
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "Queue is empty")
		return
	}

	d.changeQueueMessage(DISPLAY_QUEUE, callingPlayer)
}

func (d *Delegator) handleQueuePop() bool {
	queueLength := d.currentQueue.GetQueueLength()
	popLength := d.currentQueue.GetPopLength()

	return queueLength == popLength
}

func (d *Delegator) handleMatchOver() {

	winnerId, err := d.getGlobalName(d.DiscordUser.Author.ID, d.DiscordUser.Member.Nick)
	if err != nil {
		log.Fatal(err)
	}
	winnerImage := d.DiscordUser.Author.AvatarURL("480")
	winnerDiscordId := d.DiscordUser.Author.ID
	strWinnerDiscordId, err := strconv.Atoi(winnerDiscordId)
	if err != nil {
		log.Fatal(err)
	}

	// Determine which repository and queue type to use based on channel
	var playerRepo domain.PlayerRepository
	var queueType QueueType
	if d.DiscordUser.ChannelID == ONEVSONECHANNELID {
		playerRepo = d.PlayerRepo1v1
		queueType = QueueType1v1
	} else {
		playerRepo = d.PlayerRepo2v2
		queueType = QueueType2v2
	}

	winningPlayer := playerRepo.Get(winnerId, strWinnerDiscordId)
	winningMatch := winningPlayer.MatchId

	if winningPlayer.MatchId == uuid.Nil {
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "You are not currently in a match.")
		return
	}

	var foundWinner bool = false
	var matchFound bool = false
	activeMatches := d.MatchRepository.GetMatches()

	for _, teams := range activeMatches[winningMatch].TeamOne {
		if teams.Id == winnerId {
			foundWinner = true
			matchFound = true

			d.adjustMmr(activeMatches[winningMatch].TeamOne, activeMatches[winningMatch].TeamTwo)
		}
	}

	if !foundWinner {
		matchFound = true

		d.adjustMmr(activeMatches[winningMatch].TeamTwo, activeMatches[winningMatch].TeamOne)
	}

	if !matchFound {
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "No Matches to report.")
		return
	}
	d.displayWinMessage(winnerId, winnerImage, queueType)
	delete(activeMatches, winningMatch)

	// Update leader roles for both queues
	d.updateLeaderRoles()
}

func (d *Delegator) adjustMmr(winningPlayers []domain.Player, losingPlayers []domain.Player) {

	var winningSum float64 = 0
	var losingSum float64 = 0
	var mmrChange float64 = 0

	for _, player := range winningPlayers {
		winningSum += player.Mmr
	}

	for _, player := range losingPlayers {
		losingSum += player.Mmr
	}

	mmrChange = math.Max(20*(1-math.Pow(10, (winningSum/400))/(math.Pow(10, (winningSum/400))+math.Pow(10, (losingSum/400)))), 1)

	// Determine which repository to use based on channel
	var playerRepo domain.PlayerRepository
	if d.DiscordUser.ChannelID == ONEVSONECHANNELID {
		playerRepo = d.PlayerRepo1v1
	} else {
		playerRepo = d.PlayerRepo2v2
	}

	for _, player := range winningPlayers {
		player.Mmr += mmrChange
		player.NumWins++
		player.MatchId = uuid.Nil
		playerRepo.Update(player)
	}

	for _, player := range losingPlayers {
		player.Mmr -= mmrChange
		player.NumLosses++
		player.MatchId = uuid.Nil
		playerRepo.Update(player)
	}

}

func (d *Delegator) handleLobbyReady() {
	activeMatches := d.MatchRepository.GetMatches()

	if len(activeMatches) == 0 {
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "No Active Matches")
		return
	}

	var team1 string
	var team2 string

	for _, v := range activeMatches {
		stringifiedTeamOne := []string{}
		stringifiedTeamTwo := []string{}

		for _, player := range v.TeamOne {
			stringifiedTeamOne = append(stringifiedTeamOne, player.MentionName)
			stringifiedTeamOne = append(stringifiedTeamOne, " [")
			stringifiedTeamOne = append(stringifiedTeamOne, strconv.Itoa(int(math.Round(player.Mmr))))
			stringifiedTeamOne = append(stringifiedTeamOne, "]\n")
		}
		for _, player := range v.TeamTwo {
			stringifiedTeamTwo = append(stringifiedTeamTwo, player.MentionName)
			stringifiedTeamTwo = append(stringifiedTeamTwo, " [")
			stringifiedTeamTwo = append(stringifiedTeamTwo, strconv.Itoa(int(math.Round(player.Mmr))))
			stringifiedTeamTwo = append(stringifiedTeamTwo, "]\n")
		}

		team1 = strings.Join(stringifiedTeamOne, "")
		team2 = strings.Join(stringifiedTeamTwo, "")

	}

	// Determine field names based on queue type
	var field1Name, field2Name string
	if d.currentQueueType == QueueType1v1 {
		field1Name = "-Player 1-"
		field2Name = "-Player 2-"
	} else {
		field1Name = "-Team 1-"
		field2Name = "-Team 2-"
	}

	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0xff0000, // Red
		Description: "The following teams will now play:",
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   field1Name,
				Value:  team1,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   field2Name,
				Value:  team2,
				Inline: true,
			},
			&discordgo.MessageEmbedField{
				Name:   "Check out the leaderboard here:",
				Value:  "https://versusbot.netlify.app",
				Inline: false,
			},
		},
		//Image: &discordgo.MessageEmbedImage{
		//	URL: LOGO_URL1,
		//},
		//Thumbnail: &discordgo.MessageEmbedThumbnail{
		//	URL: LOGO_URL2,
		//},
		Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
		Title:     "Queue popped, lobby is now ready!",
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Type `!help` for a list of commands",
			IconURL: ICON_URL,
		},
	}
	d.Session.ChannelMessageSendEmbed(d.DiscordUser.ChannelID, embed)

	// Ping the correct king role based on queue type
	var kingRoleID string
	if d.currentQueueType == QueueType1v1 {
		kingRoleID = ROLEID_1V1
	} else {
		kingRoleID = ROLEID_2V2
	}
	d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, fmt.Sprintf("<@&%s> a queue has popped! Join the next queue to defend your title.", kingRoleID))
}

func (d *Delegator) changeQueueMessage(messageConst int, player domain.Player) {

	queueLength := d.currentQueue.GetQueueLength()
	commands := []string{}
	q := "**!q**"
	qDesc := "Join the queue.\n"
	leave := "**!leave**"
	leaveDesc := "Leave the queue.\n"
	report := "**!report win**"
	reportDesc := "Report a match win.\n"
	status := "**!status**"
	statusDesc := "List the players in the queue.\n"
	leaderboard := "**!leaderboard**"
	leaderboardDesc := "Displays a link to view this server's leaderboard.\n"
	active := "**!activematches**"
	activeDesc := "View all active matches (matches with no report yet).\n"
	clear := "**!clear**"
	clearDesc := "Clear the queue.\n"
	help := "**!help**"
	helpDesc := "This menu.\n"

	commands = append(commands, active, activeDesc, clear, clearDesc, help, helpDesc, leave, leaveDesc, report, reportDesc, status, statusDesc, q, qDesc, leaderboard, leaderboardDesc)

	var message string
	var title string
	var color int

	//could also be a switch
	if queueLength == 1 {
		color = 0x00ff00 // Green
	} else if queueLength == 2 {
		color = 0xffff00 // Yellow
	} else if queueLength == 3 {
		color = 0xffa500 // Orange
	}

	title = strconv.Itoa(queueLength) + " players are in the queue."

	if queueLength == 1 {
		title = strconv.Itoa(queueLength) + " player is in the queue."
	}

	switch messageConst {
	case PLAYER_ADD:
		message = player.MentionName + " has entered the queue."
	case PLAYER_LEFT:
		message = player.MentionName + " has left the queue."
	case PLAYER_ALREADY_IN_QUEUE:
		message = player.MentionName + " is already in the queue."
	case PLAYER_ALREADY_IN_MATCH:
		message = "Cannot queue while in a match. " + player.MentionName + " is already in a match."
		title = ""
	case DISPLAY_QUEUE:
		message = d.currentQueue.DisplayQueue()
		title = "Queue status"
	case PLAYER_NOT_IN_QUEUE:
		message = "Type !q to join the queue."
		title = "You are not currently in the queue."
	case DISPLAY_HELP_MENU:
		title = "**Help**"
		message = strings.Join(commands, "\n")
		color = 0x800080 // Purple
	default:
		return
	}

	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       color,
		Description: message,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "Check out the leaderboard here:",
				Value:  "https://versusbot.netlify.app",
				Inline: false,
			},
		},
		//Image: &discordgo.MessageEmbedImage{
		//	URL: LOGO_URL1,
		//},
		//Thumbnail: &discordgo.MessageEmbedThumbnail{
		//	URL: LOGO_URL2,
		//},
		Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
		Title:     title,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Type `!help` for a list of commands",
			IconURL: ICON_URL,
		},
	}
	d.Session.ChannelMessageSendEmbed(d.DiscordUser.ChannelID, embed)
}

func (d *Delegator) displayWinMessage(playerName string, playerImage string, queueType QueueType) {

	var title string
	if queueType == QueueType1v1 {
		title = playerName + " wins!"
	} else {
		title = playerName + "'s team wins!"
	}
	message := "Leaderboard has been updated."
	image := playerImage

	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0x00ff00, // Green
		Description: message,
		Fields: []*discordgo.MessageEmbedField{
			&discordgo.MessageEmbedField{
				Name:   "Check out the leaderboard here:",
				Value:  "https://versusbot.netlify.app",
				Inline: false,
			},
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: image,
		},
		Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
		Title:     title,
		Footer: &discordgo.MessageEmbedFooter{
			Text:    "Type `!help` for a list of commands",
			IconURL: ICON_URL,
		},
	}
	d.Session.ChannelMessageSendEmbed(d.DiscordUser.ChannelID, embed)
}

func (d *Delegator) handleDisplayMatches() {
	activeMatches := d.MatchRepository.GetMatches()

	if len(activeMatches) == 0 {
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "No Active Matches")
		return
	}

	var team1 string
	var team2 string
	var title string
	var num int

	for _, v := range activeMatches {
		stringifiedTeamOne := []string{}
		stringifiedTeamTwo := []string{}
		num++

		if num == 1 {
			title = "Current Matches"
		} else {
			title = ""
		}

		for _, player := range v.TeamOne {
			stringifiedTeamOne = append(stringifiedTeamOne, player.MentionName)
			stringifiedTeamOne = append(stringifiedTeamOne, " [")
			stringifiedTeamOne = append(stringifiedTeamOne, strconv.Itoa(int(math.Round(player.Mmr))))
			stringifiedTeamOne = append(stringifiedTeamOne, "]\n")
		}
		for _, player := range v.TeamTwo {
			stringifiedTeamTwo = append(stringifiedTeamTwo, player.MentionName)
			stringifiedTeamTwo = append(stringifiedTeamTwo, " [")
			stringifiedTeamTwo = append(stringifiedTeamTwo, strconv.Itoa(int(math.Round(player.Mmr))))
			stringifiedTeamTwo = append(stringifiedTeamTwo, "]\n")

		}
		//stringifiedTeamOne = stringifiedTeamOne[:len(stringifiedTeamOne)-1]
		//stringifiedTeamOne = append(stringifiedTeamOne, "]")
		//stringifiedTeamTwo = stringifiedTeamTwo[:len(stringifiedTeamTwo)-1]
		//stringifiedTeamTwo = append(stringifiedTeamTwo, "]")

		team1 = strings.Join(stringifiedTeamOne, "")
		team2 = strings.Join(stringifiedTeamTwo, "")

		embed := &discordgo.MessageEmbed{
			Author:      &discordgo.MessageEmbedAuthor{},
			Color:       0x00ff00, // Green
			Description: "Match ID: " + v.MatchUid.String(),
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "-Team 1-",
					Value:  team1,
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "-Team 2-",
					Value:  team2,
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Check out the leaderboard here:",
					Value:  "https://versusbot.netlify.app",
					Inline: false,
				},
			},
			//Image: &discordgo.MessageEmbedImage{
			//URL: LOGO_URL1,
			//},
			//Thumbnail: &discordgo.MessageEmbedThumbnail{
			//	URL: LOGO_URL2,
			//},
			Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
			Title:     title,
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Type `!help` for a list of commands",
				// IconURL: ICON_URL,
			},
		}

		d.Session.ChannelMessageSendEmbed(d.DiscordUser.ChannelID, embed)
	}

}

func (d *Delegator) handleClearQueue() {

	authorID := d.DiscordUser.Author.ID
	queueLength := d.currentQueue.GetQueueLength()

	if authorID == "189579878448889856" {
		for queueLength > 0 {
			d.currentQueue.Dequeue()
			queueLength = d.currentQueue.GetQueueLength()
		}
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "Queue has been cleared.")
	} else {
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "You do not have permission to execute this command.")
	}

}

func (d *Delegator) handleDisplayHelp() {
	incomingDiscordId := d.DiscordUser.Author.ID
	incomingId := d.DiscordUser.Author.String()
	strIncomingDiscordId, err := strconv.Atoi(incomingDiscordId)
	if err != nil {
		log.Fatal(err)
	}
	mention := d.DiscordUser.Author.Mention()
	prospectivePlayer := d.currentPlayerRepo.Get(incomingId, strIncomingDiscordId)
	prospectivePlayer.MentionName = mention

	d.changeQueueMessage(DISPLAY_HELP_MENU, prospectivePlayer)

}

func (d *Delegator) handleDisplayLeaderboard() {
	d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "Leaderboard for this server can be found at https://versusbot.netlify.app")
}

// updateLeaderRoles checks both 1v1 and 2v2 leaderboards and assigns the appropriate king roles
func (d *Delegator) updateLeaderRoles() {
	// Update 2v2 leader role
	d.updateLeaderRoleForQueue(d.PlayerRepo2v2, ROLEID_2V2)

	// Update 1v1 leader role
	d.updateLeaderRoleForQueue(d.PlayerRepo1v1, ROLEID_1V1)
}

// updateLeaderRoleForQueue gets the current leader from a queue and assigns/removes the role
func (d *Delegator) updateLeaderRoleForQueue(playerRepo domain.PlayerRepository, roleID string) {
	currentLeader := playerRepo.GetLeader()
	if currentLeader == 0 {
		// No leader found, just remove role from anyone who has it
		d.removeRoleFromAllMembers(roleID)
		return
	}

	strCurrentLeader := strconv.Itoa(currentLeader)

	// Get all guild members to find who currently has this role
	guildMembers, err := d.Session.GuildMembers(GUILDID, "", 1000)
	if err != nil {
		log.Printf("Error fetching guild members: %v", err)
		// Still try to add role to new leader even if we can't find old one
		d.Session.GuildMemberRoleAdd(GUILDID, strCurrentLeader, roleID)
		return
	}

	// Find who currently has this role
	var oldLeaderID string = ""
	for _, member := range guildMembers {
		for _, role := range member.Roles {
			if role == roleID {
				oldLeaderID = member.User.ID
				break
			}
		}
		if oldLeaderID != "" {
			break
		}
	}

	// Add role to new leader
	err = d.Session.GuildMemberRoleAdd(GUILDID, strCurrentLeader, roleID)
	if err != nil {
		log.Printf("Error adding role to leader: %v", err)
	}

	// Remove role from old leader if different
	if oldLeaderID != "" && oldLeaderID != strCurrentLeader {
		err = d.Session.GuildMemberRoleRemove(GUILDID, oldLeaderID, roleID)
		if err != nil {
			log.Printf("Error removing role from old leader: %v", err)
		}
	}
}

// removeRoleFromAllMembers removes a role from all members who have it (used when no leader exists)
func (d *Delegator) removeRoleFromAllMembers(roleID string) {
	guildMembers, err := d.Session.GuildMembers(GUILDID, "", 1000)
	if err != nil {
		log.Printf("Error fetching guild members: %v", err)
		return
	}

	for _, member := range guildMembers {
		for _, role := range member.Roles {
			if role == roleID {
				err = d.Session.GuildMemberRoleRemove(GUILDID, member.User.ID, roleID)
				if err != nil {
					log.Printf("Error removing role from member %s: %v", member.User.ID, err)
				}
				break
			}
		}
	}
}

func (d *Delegator) getGlobalName(discordID string, nickName string) (string, error) {
	// Discord API endpoint for fetching user information
	url := "https://discord.com/api/users/" + discordID

	// Create a new request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}

	// Set the authorization header to your bot token
	botToken := os.Getenv("TOKEN")

	if botToken == "" {
		err := errors.New("no token found")
		log.Fatal(err)
	}
	req.Header.Set("Authorization", "Bot "+botToken)
	req.Header.Set("Content-Type", "application/json")

	// Create a new HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var user DiscordUser
	err = json.Unmarshal([]byte(body), &user)
	if err != nil {
		panic(err)
	}

	if nickName == "" {
		return user.GlobalName, nil
	} else {
		return nickName, nil
	}
}

func (d *Delegator) specialMattFunction() {
	d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "Matt is a dingus.")
}
