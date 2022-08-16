package domain

import "strings"

type Player struct {
	Id          string
	DisplayName string
	MatchId     string
	NumWins     int
	NumLosses   int
	Mmr         float32
	IsInGame    bool
	IsAdmin     bool
}

type PlayerRepository interface {
	Store(player Player)
	Get(id string) Player
	Remove(id string)
}

func NewPlayer(id string) *Player {
	p := Player{
		Id:          id,
		DisplayName: strings.Split(id, "#")[0],
		MatchId:     id,
		NumWins:     0,
		NumLosses:   0,
		Mmr:         1000,
		IsInGame:    false,
		IsAdmin:     false,
	}

	return &p
}
