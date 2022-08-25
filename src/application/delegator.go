package application

import (
	"math"
	"strconv"
	"time"

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

const (
	PlayerAdd            int = 0
	PlayerLeft           int = 1
	PlayerShow           int = 3
	PlayerAlreadyInQueue int = 4
	PlayerAlreadyInMatch int = 5
)

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

	if strings.Contains(d.command, REPORT_WIN) {
		d.command = REPORT_WIN
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
	case MATT:
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "Matt is a dingus.")
	case DISPLAY_MATCHES:
		d.handleDisplayMatches()
	default:
		return
	}
}

// Will check the DB or memory-implementation for a player
// If no player exists, makes a new one and returns it
func (d Delegator) fetchPlayer() domain.Player {
	incomingId := d.DiscordUser.Author.String()
	prospectivePlayer := d.PlayerRepository.Get(incomingId)

	return prospectivePlayer
}

func (d *Delegator) handleEnterQueue() {

	prospectivePlayer := d.fetchPlayer()
	prospectivePlayer.MentionName = d.DiscordUser.Author.Mention()

	if d.queue.PlayerInQueue(prospectivePlayer) {
		d.changeQueueMessage(PlayerAlreadyInQueue)
		return
	}

	if prospectivePlayer.MatchId != uuid.Nil {
		d.changeQueueMessage(PlayerAlreadyInMatch)
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
	d.changeQueueMessage(PlayerAdd)
}

func (d *Delegator) handleLeaveQueue() {
	incomingId := d.DiscordUser.Author.String()
	prospectivePlayer := d.PlayerRepository.Get(incomingId)

	if !d.queue.PlayerInQueue(prospectivePlayer) {
		// Player isn't in queue, exit
		return
	}

	playerSuccessfullyRemoved := d.queue.LeaveQueue(prospectivePlayer)

	if playerSuccessfullyRemoved {
		d.changeQueueMessage(PlayerLeft)
	}
}

func (d Delegator) handleDisplayQueue() {
	presentationqueue := d.queue.DisplayQueue()

	if presentationqueue == "" {
		d.Session.ChannelMessageSend(d.DiscordUser.ChannelID, "Queue is empty")
		return
	}

	d.displayQueueMessage()
}

func (d *Delegator) handleQueuePop() bool {
	queueLength := d.queue.GetQueueLength()
	popLength := d.queue.GetPopLength()

	return queueLength == popLength
}

func (d *Delegator) handleMatchOver() {

	winnerId := d.DiscordUser.Author.String()
	winningPlayer := d.PlayerRepository.Get(winnerId)
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
	d.displayWinMessage()
	delete(activeMatches, winningMatch)
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

	mmrChange = math.Max(20*(1-math.Pow(10, (winningSum/400))/((math.Pow(10, winningSum/400))+math.Pow(10, (losingSum/400)))), 1)

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
		},
		Image: &discordgo.MessageEmbedImage{
			URL: "",
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "",
		},
		Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
		Title:     "Queue popped, lobby is now ready!",
	}
	d.Session.ChannelMessageSendEmbed(d.DiscordUser.ChannelID, embed)
}

func (d *Delegator) changeQueueMessage(messageConst int) {

	queueLength := d.queue.GetQueueLength()
	var message string
	var queueStatus string
	var color int

	//could also be a switch
	if queueLength == 1 {
		color = 0x00ff00 // Green
	} else if queueLength == 2 {
		color = 0xffff00 // Yellow
	} else if queueLength == 3 {
		color = 0xffa500 // Orange
	}

	queueStatus = strconv.Itoa(queueLength) + " players are in the queue."

	switch messageConst {
	case PlayerAdd:
		message = d.DiscordUser.Author.Mention() + " has entered the queue."
	case PlayerLeft:
		message = d.DiscordUser.Author.Mention() + " has left the queue."
	case PlayerAlreadyInQueue:
		message = d.DiscordUser.Author.Mention() + " is already in the queue."
	case PlayerAlreadyInMatch:
		message = "Cannot queue while in a match. " + d.DiscordUser.Author.Mention() + " is already in a match."
		queueStatus = ""
	default:
		return
	}

	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       color,
		Description: message,
		Image: &discordgo.MessageEmbedImage{
			URL: "",
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "",
		},
		Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
		Title:     queueStatus,
	}
	d.Session.ChannelMessageSendEmbed(d.DiscordUser.ChannelID, embed)
}

func (d *Delegator) displayQueueMessage() {

	var message string = d.queue.DisplayQueue()

	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0x00ff00, // Green
		Description: message,
		Fields:      []*discordgo.MessageEmbedField{},
		Image: &discordgo.MessageEmbedImage{
			URL: "",
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: "",
		},
		Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
		Title:     "Queue status",
	}
	d.Session.ChannelMessageSendEmbed(d.DiscordUser.ChannelID, embed)
}

func (d *Delegator) displayWinMessage() {

	title := d.DiscordUser.Author.Username + "'s team wins!"
	message := "Leaderboard has been updated."
	image := d.DiscordUser.Author.AvatarURL("240")

	embed := &discordgo.MessageEmbed{
		Author:      &discordgo.MessageEmbedAuthor{},
		Color:       0x00ff00, // Green
		Description: message,
		Fields:      []*discordgo.MessageEmbedField{},
		Image: &discordgo.MessageEmbedImage{
			URL: "",
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: image,
		},
		Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
		Title:     title,
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
			},
			Image: &discordgo.MessageEmbedImage{
				URL: "",
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: "",
			},
			Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
			Title:     title,
		}

		d.Session.ChannelMessageSendEmbed(d.DiscordUser.ChannelID, embed)
	}

}
