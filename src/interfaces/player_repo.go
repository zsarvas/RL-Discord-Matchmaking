package interfaces

import (
	"fmt"

	"github.com/zsarvas/RL-Discord-Matchmaking/domain"
)

type PlayerDataHandler interface {
	Add(newPlayer domain.Player)
	Remove(id string)
	GetById(id string) domain.Player
}

type PlayerDataRepo struct {
	dbHandler PlayerDataHandler
	dbStuff   DbHandler
}

type PlayerRepo PlayerDataRepo

type DbHandler interface {
	Execute(statement string)
	Query(statement string) Row
}

type Row interface {
	Scan(dest ...interface{})
	Next() bool
}

func NewPlayerRepo(repoHandler PlayerDataHandler) *PlayerRepo {
	dbPlayerRepo := new(PlayerRepo)
	dbPlayerRepo.dbHandler = repoHandler

	return dbPlayerRepo
}

func (repo *PlayerRepo) Store(player domain.Player) {
	repo.dbHandler.Add(player)
	repo.dbStuff.Execute(fmt.Sprintf(`INSERT INTO players (id, display_name)VALUES ('%v', '%v')`, player.Id, player.DisplayName))
}

func (repo *PlayerRepo) Get(playerId string) domain.Player {
	foundPlayer := repo.dbHandler.GetById(playerId)

	return foundPlayer
}

func (repo *PlayerRepo) Remove(id string) {
	repo.dbHandler.Remove(id)
}
