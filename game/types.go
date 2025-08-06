package game

type Player struct {
	ID string `json:"id"`
}

type GameState struct {
	Players map[string]*Player `json:"players"`
}
