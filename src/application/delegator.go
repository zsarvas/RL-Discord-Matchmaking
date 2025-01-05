package application

import (
	"errors"
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
	Session          *discordgo.Session
	DiscordUser      *discordgo.MessageCreate
	queue            *domain.Queue
	PlayerRepository domain.PlayerRepository
	MatchRepository  MatchRepository
	command          string
}

type DiscordUser struct {
	ID                   string  `json:"id"`
	Username             string  `json:"username"`
	Avatar               string  `json:"avatar"`
	Discriminator        string  `json:"discriminator"`
	PublicFlags          int     `json:"public_flags"`
	PremiumType          int     `json:"premium_type"`
	Flags                int     `json:"flags"`
	Banner               *string `json:"banner"`       // Pointer to handle null value
	AccentColor          *int    `json:"accent_color"` // Pointer to handle null value
	GlobalName           string  `json:"global_name"`
	AvatarDecorationData *string `json:"avatar_decoration_data"` // Pointer to handle null value
	BannerColor          *string `json:"banner_color"`           // Pointer to handle null value
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
	ICON_URL                string = ""
	FOURMANSCHANNELID       string = "1011004892418166877"
)

var playerTimers = make(map[domain.Player]*time.Timer)

func NewDelegator(playerRepo domain.PlayerRepository, matchRepo MatchRepository) *Delegator {
	// could possible move this queue out of the 'constructor'
	newQueue := domain.NewQueue(4)

	cd := &Delegator{
		PlayerRepository: playerRepo,
		MatchRepository:  matchRepo,
		queue:            newQueue,
	}

	return cd
}

func (d *Delegator) InitiateDelegator(s *discordgo.Session, m *discordgo.MessageCreate) {
	d.Session = s
	d.DiscordUser = m
	d.command = strings.ToUpper(m.Content)

	if m.ChannelID != FOURMANSCHANNELID {
		m.Content = ""
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
	prospectivePlayer := d.PlayerRepository.Get(globalName, strIncomingDiscordId)
	prospectivePlayer.MentionName = mention
	prospectivePlayer.Id = globalName

	return prospectivePlayer
}

func (d *Delegator) handleEnterQueue() {

	prospectivePlayer := d.fetchPlayer()
	d.startTimeoutTimer(prospectivePlayer)

	if d.queue.PlayerInQueue(prospectivePlayer) {
		d.changeQueueMessage(PLAYER_ALREADY_IN_QUEUE, prospectivePlayer)
		return
	}

	if prospectivePlayer.MatchId != uuid.Nil {
		d.changeQueueMessage(PLAYER_ALREADY_IN_MATCH, prospectivePlayer)
		return
	}

	d.queue.Enqueue(prospectivePlayer)

	queueIsPopping := d.handleQueuePop()

	if queueIsPopping {
		d.queue.RandomizeQueue()
		match := Match{
			TeamOne: []domain.Player{d.queue.Dequeue(), d.queue.Dequeue()},
			TeamTwo: []domain.Player{d.queue.Dequeue(), d.queue.Dequeue()},
		}

		matchId := d.MatchRepository.Add(match)

		for _, player := range match.TeamOne {
			player.MatchId = matchId
			d.PlayerRepository.SetMatch(player)
		}

		for _, player := range match.TeamTwo {
			player.MatchId = matchId
			d.PlayerRepository.SetMatch(player)
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
	playerTimers[player] = time.AfterFunc(20*time.Minute, func() {
		if d.queue.PlayerInQueue(player) {
			d.Session.ChannelMessageSend(FOURMANSCHANNELID, player.MentionName+" has been timed out from the queue.")
			d.queue.LeaveQueue(player)
			d.changeQueueMessage(PLAYER_LEFT, player)
		}
	})
}

func (d *Delegator) stopTimeoutTimer(player domain.Player) {
	if timer, ok := playerTimers[player]; ok {
		timer.Stop()
		delete(playerTimers, player)
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
	prospectivePlayer := d.PlayerRepository.Get(globalName, strIncomingDiscordId)
	prospectivePlayer.MentionName = mention
	prospectivePlayer.Id = globalName

	if !d.queue.PlayerInQueue(prospectivePlayer) {
		// Player isn't in queue, exit
		d.changeQueueMessage(PLAYER_NOT_IN_QUEUE, prospectivePlayer)
		return
	}

	playerSuccessfullyRemoved := d.queue.LeaveQueue(prospectivePlayer)

	if playerSuccessfullyRemoved {
		d.changeQueueMessage(PLAYER_LEFT, prospectivePlayer)
	}

	d.stopTimeoutTimer(prospectivePlayer)
}

func (d Delegator) handleDisplayQueue() {
	incomingDiscordId := d.DiscordUser.Author.ID
	presentationqueue := d.queue.DisplayQueue()
	strIncomingDiscordId, err := strconv.Atoi(incomingDiscordId)
	if err != nil {
		log.Fatal(err)
	}
	globalName, err := d.getGlobalName(d.DiscordUser.Author.ID, d.DiscordUser.Member.Nick)
	if err != nil {
		log.Fatal(err)
	}
	callingPlayer := d.PlayerRepository.Get(globalName, strIncomingDiscordId)

	if presentationqueue == "" {
		d.Session.ChannelMessageSend(FOURMANSCHANNELID, "Queue is empty")
		return
	}

	d.changeQueueMessage(DISPLAY_QUEUE, callingPlayer)
}

func (d *Delegator) handleQueuePop() bool {
	queueLength := d.queue.GetQueueLength()
	popLength := d.queue.GetPopLength()

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

	winningPlayer := d.PlayerRepository.Get(winnerId, strWinnerDiscordId)
	winningMatch := winningPlayer.MatchId

	oldLeader := d.PlayerRepository.GetLeader()

	if winningPlayer.MatchId == uuid.Nil {
		d.Session.ChannelMessageSend(FOURMANSCHANNELID, "You are not currently in a match.")
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
		d.Session.ChannelMessageSend(FOURMANSCHANNELID, "No Matches to report.")
		return
	}
	d.displayWinMessage(winnerId, winnerImage)
	delete(activeMatches, winningMatch)
	leader := d.PlayerRepository.GetLeader()
	strLeader := strconv.Itoa(leader)
	strOldLeader := strconv.Itoa(oldLeader)
	d.handleLeaderRole(strLeader, strOldLeader)
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

	for _, player := range winningPlayers {
		player.Mmr += mmrChange
		player.NumWins++
		player.MatchId = uuid.Nil
		d.PlayerRepository.Update(player)
	}

	for _, player := range losingPlayers {
		player.Mmr -= mmrChange
		player.NumLosses++
		player.MatchId = uuid.Nil
		d.PlayerRepository.Update(player)
	}

}

func (d *Delegator) handleLobbyReady() {
	activeMatches := d.MatchRepository.GetMatches()

	if len(activeMatches) == 0 {
		d.Session.ChannelMessageSend(FOURMANSCHANNELID, "No Active Matches")
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

	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0xff0000, // Red
		Description: "The following teams will now play:",
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
		//	URL: LOGO_URL1,
		//},
		//Thumbnail: &discordgo.MessageEmbedThumbnail{
		//	URL: LOGO_URL2,
		//},
		Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
		Title:     "Queue popped, lobby is now ready!",
		Footer: &discordgo.MessageEmbedFooter{
			Text: "Powered by 2CDs",
			// IconURL: ICON_URL,
		},
	}
	d.Session.ChannelMessageSendEmbed(FOURMANSCHANNELID, embed)
	d.Session.ChannelMessageSend(FOURMANSCHANNELID, "<@&1028789594277302302> a queue has popped!  Join the next queue to defend your title.")
}

func (d *Delegator) changeQueueMessage(messageConst int, player domain.Player) {

	queueLength := d.queue.GetQueueLength()
	commands := []string{}
	active := "**!activematches**"
	activeDesc := "View all active matches (matches with no report yet).\n"
	clear := "**!clear**"
	clearDesc := "Clear the queue.\n"
	help := "**!help**"
	helpDesc := "This menu.\n"
	leave := "**!leave**"
	leaveDesc := "Leave the queue.\n"
	report := "**!report win**"
	reportDesc := "Report a match win.\n"
	status := "**!status**"
	statusDesc := "List the players in the queue.\n"
	q := "**!q**"
	qDesc := "Join the queue.\n"
	leaderboard := "**!leaderboard**"
	leaderboardDesc := "Displays a link to view this server's leaderboard.\n"

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
		message = d.queue.DisplayQueue()
		title = "Queue status"
	case PLAYER_NOT_IN_QUEUE:
		message = "Type !q to join the queue."
		title = "You are not currently in the queue."
	case DISPLAY_HELP_MENU:
		title = "**Help**"
		message = strings.Join(commands, "\n")
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
			Text: "Powered by 2CDs",
			// IconURL: ICON_URL,
		},
	}
	d.Session.ChannelMessageSendEmbed(FOURMANSCHANNELID, embed)
}

func (d *Delegator) displayWinMessage(playerName string, playerImage string) {

	title := playerName + "'s team wins!"
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
			Text: "Powered by 2CDs",
			// IconURL: ICON_URL,
		},
	}
	d.Session.ChannelMessageSendEmbed(FOURMANSCHANNELID, embed)
}

func (d *Delegator) handleDisplayMatches() {
	activeMatches := d.MatchRepository.GetMatches()

	if len(activeMatches) == 0 {
		d.Session.ChannelMessageSend(FOURMANSCHANNELID, "No Active Matches")
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
				Text: "Powered by 2CDs",
				// IconURL: ICON_URL,
			},
		}

		d.Session.ChannelMessageSendEmbed(FOURMANSCHANNELID, embed)
	}

}

func (d *Delegator) handleClearQueue() {

	authorID := d.DiscordUser.Author.ID
	queueLength := d.queue.GetQueueLength()

	if authorID == "189579878448889856" {
		for queueLength > 0 {
			d.queue.Dequeue()
			queueLength = d.queue.GetQueueLength()
		}
		d.Session.ChannelMessageSend(FOURMANSCHANNELID, "Queue has been cleared.")
	} else {
		d.Session.ChannelMessageSend(FOURMANSCHANNELID, "You do not have permission to execute this command.")
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
	prospectivePlayer := d.PlayerRepository.Get(incomingId, strIncomingDiscordId)
	prospectivePlayer.MentionName = mention

	d.changeQueueMessage(DISPLAY_HELP_MENU, prospectivePlayer)

}

func (d *Delegator) handleDisplayLeaderboard() {
	d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "Leaderboard for this server can be found at https://versusbot.netlify.app")
}

func (d *Delegator) handleLeaderRole(leader string, oldLeader string) {

	d.Session.GuildMemberRoleAdd("189628012604555265", leader, "1028789594277302302")
	if leader != oldLeader {
		d.Session.GuildMemberRoleRemove("189628012604555265", oldLeader, "1028789594277302302")
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
