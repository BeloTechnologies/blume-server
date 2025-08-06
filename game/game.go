package game

func NewGameState() *GameState {
	return &GameState{
		Players: make(map[string]*Player),
	}
}

func (gs *GameState) AddPlayer(player *Player) {
	gs.Players[player.ID] = player
}

func (gs *GameState) RemovePlayer(playerID string) {
	delete(gs.Players, playerID)
}
