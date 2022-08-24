package domain

import (
	"strings"

	"github.com/google/uuid"
)

type Player struct {
	Id          string
	MentionName string
	DisplayName string
	MatchId     uuid.UUID
	NumWins     int
	NumLosses   int
	Mmr         float64
	IsInGame    bool
	IsAdmin     bool
}

type PlayerRepository interface {
	Store(player Player)
	Get(id string) Player
	Update(player Player)
	SetMatch(player Player)
}

func NewPlayer(id string) *Player {
	p := Player{
		Id:          id,
		MentionName: "",
		DisplayName: strings.Split(id, "#")[0],
		MatchId:     uuid.Nil,
		NumWins:     0,
		NumLosses:   0,
		Mmr:         1000,
		IsInGame:    false,
		IsAdmin:     false,
	}

	return &p
}
