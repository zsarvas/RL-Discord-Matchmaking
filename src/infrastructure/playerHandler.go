package infrastructure

import "github.com/zsarvas/RL-Discord-Matchmaking/domain"

type PlayerHandler struct {
	players map[string]domain.Player
}

func (handler *PlayerHandler) Add(newPlayer domain.Player) {
	handler.players[newPlayer.Id] = newPlayer
}

func (handler *PlayerHandler) Remove(id string) {
	delete(handler.players, id)
}

func (handler *PlayerHandler) GetById(id string) domain.Player {
	player, playerExists := handler.players[id]

	if !playerExists {
		newPlayerToAdd := domain.NewPlayer(id)
		handler.Add(*newPlayerToAdd)

		return *newPlayerToAdd
	}

	return player
}

func NewPlayerHandler() *PlayerHandler {
	plyrHandler := &PlayerHandler{
		players: make(map[string]domain.Player),
	}

	return plyrHandler
}
