package infrastructure

import (
	"database/sql"

	"fmt"
	"log"

	_ "github.com/lib/pq"

	"github.com/google/uuid"

	"github.com/zsarvas/RL-Discord-Matchmaking/domain"
)

type PlayerHandler struct {
	Conn *sql.DB
}

func (handler *PlayerHandler) Add(newPlayer domain.Player) {
	handler.Conn.Exec(fmt.Sprintf(`INSERT INTO rocketleague ("Name", "MMR", "Wins", "Losses", "MatchUID", "DiscordId") VALUES ('%v', '%f', '%d', '%d', '%s', '%d');`, newPlayer.Id, newPlayer.Mmr, newPlayer.NumWins, newPlayer.NumLosses, newPlayer.MatchId, newPlayer.DiscordId))
}

func (handler *PlayerHandler) GetById(id string, uniqueId int) domain.Player {
	record, err := handler.Conn.Query(`SELECT * FROM rocketleague WHERE "DiscordId" = $1;`, uniqueId)

	if err != nil {
		log.Fatal(err)
	}

	var index int
	var name string
	var mmr float64
	var numWins int
	var numLosses int
	var matchId uuid.UUID
	var discordId int

	for record.Next() {
		record.Scan(&index, &name, &mmr, &numWins, &numLosses, &matchId, &discordId)
	}

	if uniqueId != discordId {
		newPlayer := domain.NewPlayer(id, uniqueId)
		handler.Add(*newPlayer)
		return *newPlayer
	} else {
		foundPlayer := new(domain.Player)
		foundPlayer.Id = id
		foundPlayer.NumWins = numWins
		foundPlayer.NumLosses = numLosses
		foundPlayer.Mmr = mmr
		foundPlayer.MatchId = matchId
		foundPlayer.DiscordId = uniqueId

		return *foundPlayer
	}
}

func NewPlayerHandler(connStr string) *PlayerHandler {

	connection, err := sql.Open("postgres", connStr)

	if err != nil {
		fmt.Println(err)
	}

	plyrHandler := &PlayerHandler{Conn: connection}

	return plyrHandler
}

func (handler *PlayerHandler) UpdatePlayer(player domain.Player) {
	handler.Conn.Exec(`UPDATE rocketleague SET "MMR" = $1, "Wins" = $2, "Losses" = $3, "MatchUID" = $4, "Name" = $5 WHERE "DiscordId" = $6;`, player.Mmr, player.NumWins, player.NumLosses, player.MatchId, player.Id, player.DiscordId)
}

func (handler *PlayerHandler) SetMatchId(player domain.Player) {
	handler.Conn.Exec(`UPDATE rocketleague SET "MatchUID" = $1 WHERE "DiscordId" = $2;`, player.MatchId, player.DiscordId)
}

func (handler *PlayerHandler) GetLead() int {
	record, err := handler.Conn.Query(`SELECT DISTINCT ON ("MMR") "id", "Name", "MMR", "Wins", "Losses", "MatchUID", "DiscordId" FROM rocketleague ORDER BY "MMR" DESC`)

	if err != nil {
		log.Fatal(err)
	}
	var index int
	var name string
	var mmr float64
	var numWins int
	var numLosses int
	var matchId uuid.UUID
	var discordId int

	for record.Next() {
		record.Scan(&index, &name, &mmr, &numWins, &numLosses, &matchId, &discordId)
		return discordId
	}

	return 0
}
