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
	DiscordId   int
}

type PlayerRepository interface {
	Store(player Player)
	Get(id string, uniqueId int) Player
	Update(player Player)
	SetMatch(player Player)
	GetLeader() int
}

func NewPlayer(id string, uniqueId int) *Player {
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
		DiscordId:   uniqueId,
	}

	return &p
}
