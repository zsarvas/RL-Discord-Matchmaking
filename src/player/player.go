package player

import (
	"strings"
)

type player struct {
	id          string
	displayName string
	matchId     string
	numWins     int
	numLosses   int
	mmr         float32
	inGame      bool
}

func newPlayer(id string) *player {

	p := player{id: id}
	p.displayName = strings.Split(id, "#")[0]
	return &p
}
