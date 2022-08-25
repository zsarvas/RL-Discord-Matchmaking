package interfaces

import (
	"github.com/google/uuid"
	"github.com/zsarvas/RL-Discord-Matchmaking/application"
)

type MatchDataHandler interface {
	AddMatch(match application.Match) uuid.UUID
	GetActiveMatches() map[uuid.UUID]application.Match
}

type MatchDataRepo struct {
	dataHandler MatchDataHandler
}

type MatchRepo MatchDataRepo

func NewMatchDataRepo(repoHandler MatchDataHandler) *MatchRepo {
	matchRepo := new(MatchRepo)
	matchRepo.dataHandler = repoHandler

	return matchRepo
}

func (repo *MatchRepo) Add(match application.Match) uuid.UUID {
	return repo.dataHandler.AddMatch(match)
}

func (repo *MatchRepo) GetMatches() map[uuid.UUID]application.Match {
	return repo.dataHandler.GetActiveMatches()
}
