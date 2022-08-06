package application

import "github.com/zsarvas/RL-Discord-Matchmaking/domain"

type Team = []domain.Player

type Match struct {
	TeamOne Team
	TeamTwo Team
}

type MatchHolder struct {
	Match         Match
	Winners       Team
	ActiveMatches map[string]Match
}

type MatchRepository interface {
	AddMatch(match MatchHolder)
}
