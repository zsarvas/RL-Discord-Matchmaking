package infrastructure

import (
	"github.com/google/uuid"
	"github.com/zsarvas/RL-Discord-Matchmaking/application"
)

type MatchHandler struct {
	ActiveMatches map[uuid.UUID]application.Match
}

func NewMatchHandler() *MatchHandler {
	matchHandler := &MatchHandler{
		ActiveMatches: make(map[uuid.UUID]application.Match),
	}

	return matchHandler
}

func (mh *MatchHandler) AddMatch(match application.Match) uuid.UUID {
	createdUuid := uuid.New()

	mh.ActiveMatches[createdUuid] = match
	match.MatchUid = createdUuid

	return match.MatchUid
}

func (mh *MatchHandler) GetActiveMatches() map[uuid.UUID]application.Match {
	return mh.ActiveMatches
}
