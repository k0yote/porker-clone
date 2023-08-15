package p2p

type Round uint32

const (
	Dealing Round = iota
	PreFlop
	Flop
	Turn
	River
)

type GameState struct {
	isDealer bool
	Round    uint32
}

func NewGameState() *GameState {
	return &GameState{}
}

func (g *GameState) loop() {
	for {
		select {}
	}
}
