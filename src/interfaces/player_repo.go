package interfaces

import (
	"github.com/zsarvas/RL-Discord-Matchmaking/domain"
)

type PlayerDataHandler interface {
	Add(newPlayer domain.Player)
	GetById(id string, uniqueId int) domain.Player
	UpdatePlayer(player domain.Player)
	SetMatchId(player domain.Player)
	GetLead() int
}

type PlayerDataRepo struct {
	dbHandler PlayerDataHandler
}

type PlayerRepo PlayerDataRepo

func NewPlayerRepo(repoHandler PlayerDataHandler) *PlayerRepo {
	dbPlayerRepo := new(PlayerRepo)
	dbPlayerRepo.dbHandler = repoHandler

	return dbPlayerRepo
}

func (repo *PlayerRepo) Store(player domain.Player) {
	repo.dbHandler.Add(player)
}

func (repo *PlayerRepo) Get(playerId string, uniqueId int) domain.Player {
	foundPlayer := repo.dbHandler.GetById(playerId, uniqueId)

	return foundPlayer
}

func (repo *PlayerRepo) Update(player domain.Player) {
	repo.dbHandler.UpdatePlayer(player)
}

func (repo *PlayerRepo) SetMatch(player domain.Player) {
	repo.dbHandler.SetMatchId(player)
}

func (repo *PlayerRepo) GetLeader() int {
	return repo.dbHandler.GetLead()
}
