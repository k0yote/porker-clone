package p2p

type GameStatus int32

func (g GameStatus) String() string {
	switch g {
	case GameStatusWaitingForCards:
		return "Waiting FOR CARDS"
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
	GameStatusWaitingForCards GameStatus = iota
	GameStatusShuffleAndDeal
	GameStatusReceivingCards
	GameStatusDealing
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)
