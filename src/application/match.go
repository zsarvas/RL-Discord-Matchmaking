package application

import (
	"github.com/google/uuid"
	"github.com/zsarvas/RL-Discord-Matchmaking/domain"
)

type Team = []domain.Player

type Match struct {
	TeamOne Team
	TeamTwo Team
}

type MatchRepository interface {
	Add(match Match)
	GetMatches() map[uuid.UUID]Match
}
