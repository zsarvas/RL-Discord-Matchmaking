package player

type player struct {
	id        string
	matchId   string
	numWins   int
	numLosses int
	mmr       float32
	inGame    bool
}

func newPlayer(id string) *player {

	p := player{id: id}
	return &p
}
