package p2p

type GameStatus int32

func (g GameStatus) String() string {
	switch g {
	case GameStatusConnected:
		return "CONNECTED"
	case GameStatusPlayerReady:
		return "PLAYER READY"
	case GameStatusShuffleAndDeal:
		return "Shuffle And Deal"
	case GameStatusReceivingCards:
		return "Receiving CARDS"
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
	GameStatusConnected GameStatus = iota
	GameStatusPlayerReady
	GameStatusShuffleAndDeal
	GameStatusReceivingCards
	GameStatusDealing
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)
