package p2p

type GameStatus uint32

func (g GameStatus) String() string {
	switch g {
	case GameStatusWaiting:
		return "Waiting"
	case GameStatusDealing:
		return "Dealing"
	case GameStatusPreFlop:
		return "Pre-Flop"
	case GameStatusFlop:
		return "Flop"
	case GameStatusTurn:
		return "Turn"
	case GameStatusRiver:
		return "River"
	default:
		return "unknown"
	}
}

const (
	GameStatusWaiting GameStatus = iota
	GameStatusDealing
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)

type GameState struct {
	isDealer   bool
	gameStatus GameStatus
}

func NewGameState() *GameState {
	return &GameState{}
}

func (g *GameState) loop() {
	for {
		select {}
	}
}
