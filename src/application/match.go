package application

import (
	"github.com/google/uuid"
	"github.com/zsarvas/RL-Discord-Matchmaking/domain"
)

type Team = []domain.Player

type Match struct {
	TeamOne  Team
	TeamTwo  Team
	MatchUid uuid.UUID
}

type MatchRepository interface {
	Add(match Match) uuid.UUID
	GetMatches() map[uuid.UUID]Match
}
