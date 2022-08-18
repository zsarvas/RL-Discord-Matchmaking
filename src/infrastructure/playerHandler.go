package infrastructure

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zsarvas/RL-Discord-Matchmaking/domain"
)

type PlayerHandler struct {
	Conn *sql.DB
}

func (handler *PlayerHandler) Add(newPlayer domain.Player) {
	handler.Conn.Exec(fmt.Sprintf(`INSERT INTO players (Name, MMR, Wins, Losses) VALUES ('%v', '%f', '%d', '%d')`, newPlayer.Id, newPlayer.Mmr, newPlayer.NumWins, newPlayer.NumLosses))
}

func (handler *PlayerHandler) GetById(id string) domain.Player {
	record, err := handler.Conn.Query("SELECT * FROM players WHERE Name = ?", id)
	if err != nil {
		log.Fatal(err)
	}

	var index int
	var name string
	var mmr float64
	var numWins int
	var numLosses int

	for record.Next() {
		record.Scan(&index, &name, &mmr, &numWins, &numLosses)
	}

	if id != name {
		newPlayer := domain.NewPlayer(id)
		handler.Add(*newPlayer)
		return *newPlayer
	} else {
		foundPlayer := new(domain.Player)
		foundPlayer.Id = id
		foundPlayer.DisplayName = strings.Split(id, "#")[0]
		foundPlayer.NumWins = numWins
		foundPlayer.NumLosses = numLosses
		foundPlayer.Mmr = mmr

		return *foundPlayer
	}
}

func NewPlayerHandler(dbFileName string) *PlayerHandler {
	connection, err := sql.Open("sqlite3", dbFileName)

	if err != nil {
		log.Fatal(err)
	}

	pingError := connection.Ping()

	if pingError != nil {
		log.Fatal(pingError)
	}

	plyrHandler := &PlayerHandler{Conn: connection}

	createTable(connection)

	return plyrHandler
}

func createTable(db *sql.DB) {

	players_table := `CREATE TABLE IF NOT EXISTS players (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"Name" TEXT,
		"MMR" INT,
		"Wins" INT,
		"Losses" INT);`

	query, err := db.Prepare(players_table)
	if err != nil {
		log.Fatal(err)
	}
	query.Exec()
}
