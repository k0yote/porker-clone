package p2p

type PlayerAction byte

func (pa PlayerAction) String() string {
	switch pa {
	case PlayerActionFold:
		return "FOLD"
	case PlayerActionCheck:
		return "CHECK"
	case PlayerActionBet:
		return "BET"
	default:
		return "INVALID ACTION"
	}
}

const (
	PlayerActionFold PlayerAction = iota + 1
	PlayerActionCheck
	PlayerActionBet
)

type GameStatus int32

func (g GameStatus) String() string {
	switch g {
	case GameStatusConnected:
		return "CONNECTED"
	case GameStatusPlayerReady:
		return "PLAYER READY"
	case GameStatusDealing:
		return "Dealing"
	case GameStatusFolded:
		return "FOLED"
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
	GameStatusDealing
	GameStatusFolded
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)
