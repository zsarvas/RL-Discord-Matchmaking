package infrastructure

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zsarvas/RL-Discord-Matchmaking/domain"
)

type PlayerHandler struct {
	Conn *sql.DB
}

type Row interface {
	Scan(dest ...interface{})
	Next() bool
}

func (handler *PlayerHandler) Add(newPlayer domain.Player) {
	handler.Conn.Exec(fmt.Sprintf(`INSERT INTO players (UID, Name, MMR, Wins, Losses) VALUES ('%v', '%v', '%f', '%d', '%d')`, newPlayer.Id, newPlayer.DisplayName, newPlayer.Mmr, newPlayer.NumWins, newPlayer.NumLosses))
}

func (handler *PlayerHandler) Remove(id string) {
	panic(1)
}

func (handler *PlayerHandler) GetById(id string) domain.Player {

	fmt.Printf("id is : '%v' \n", id)

	record, err := handler.Conn.Query("SELECT * FROM players WHERE UID = ?", id)
	if err != nil {
		log.Fatal(err)
	}

	var dbId int
	var uid string
	var displayName string
	var mmr float64
	var numWins int
	var numLosses int

	for record.Next() {
		record.Scan(&dbId, &uid, &displayName, &mmr, &numWins, &numLosses)
	}

	if id != uid {
		newPlayer := domain.NewPlayer(id)
		handler.Add(*newPlayer)
		return *newPlayer
	} else {
		foundPlayer := new(domain.Player)
		foundPlayer.Id = id
		foundPlayer.DisplayName = displayName
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
	//showTable(connection)

	return plyrHandler
}

func createTable(db *sql.DB) {

	players_table := `CREATE TABLE IF NOT EXISTS players (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		"UID" TEXT,
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

/*func showTable(db *sql.DB) {
	rows, err := db.Query("SELECT * FROM players")
	if err != nil {
		log.Fatal(err)
	}

	var uid int
	var username string
	var mmr sql.NullString
	var wins sql.NullInt64
	var losses sql.NullInt64

	for rows.Next() {
		err = rows.Scan(&uid, &username, &mmr, &wins, &losses)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(uid)
		fmt.Println(username)
		fmt.Println(mmr)
		fmt.Println(wins)
		fmt.Println(losses)
	}
}*/
