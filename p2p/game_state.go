package p2p

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/k0yote/k0porker/deck"
	"github.com/sirupsen/logrus"
)

type GameStatus int32

func (g GameStatus) String() string {
	switch g {
	case GameStatusWaitingForCards:
		return "Waiting FOR CARDS"
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
	GameStatusReceivingCards
	GameStatusDealing
	GameStatusPreFlop
	GameStatusFlop
	GameStatusTurn
	GameStatusRiver
)

type Player struct {
	Status GameStatus
}

type GameState struct {
	listenAddr  string
	broadcastCh chan any
	isDealer    bool
	gameStatus  GameStatus

	playersWaitingForCards int32
	playersLock            sync.RWMutex
	players                map[string]*Player
}

func NewGameState(addr string, broadcastCh chan any) *GameState {
	g := &GameState{
		listenAddr:  addr,
		broadcastCh: broadcastCh,
		isDealer:    false,
		gameStatus:  GameStatusWaitingForCards,
		players:     make(map[string]*Player),
	}

	go g.loop()

	return g
}

func (g *GameState) SetStatus(s GameStatus) {
	// oldStatus := atomic.LoadInt32((*int32)(&g.gameStatus))
	atomic.StoreInt32((*int32)(&g.gameStatus), (int32)(s))
}

func (g *GameState) AddPlayerWaitingForCards() {
	atomic.AddInt32(&g.playersWaitingForCards, 1)
}

func (g *GameState) CheckNeedDealCards() {
	playersWaiting := atomic.LoadInt32(&g.playersWaitingForCards)

	if playersWaiting == int32(len(g.players)) &&
		g.isDealer &&
		g.gameStatus == GameStatusWaitingForCards {
		logrus.WithFields(logrus.Fields{
			"addr": g.listenAddr,
		}).Info("need to deal cards")

		g.DealCards()
	}
}

func (g *GameState) DealCards() {
	g.broadcastCh <- MessageCards{Deck: deck.New()}
}

func (g *GameState) SetPlayerStatus(addr string, status GameStatus) {
	player, ok := g.players[addr]
	if !ok {
		panic("player not found although it should exist")
	}

	player.Status = status

	g.CheckNeedDealCards()
}

func (g *GameState) LenPlayersConnectedWithLock() int {
	g.playersLock.RLock()
	defer g.playersLock.RUnlock()

	return len(g.players)
}

func (g *GameState) AddPlayer(addr string, status GameStatus) {
	g.playersLock.Lock()
	defer g.playersLock.Unlock()

	if status == GameStatusWaitingForCards {
		g.AddPlayerWaitingForCards()
	}

	g.players[addr] = new(Player)

	g.SetPlayerStatus(addr, status)

	logrus.WithFields(logrus.Fields{
		"addr":   addr,
		"status": status,
	}).Info("new player joined")
}

func (g *GameState) loop() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-ticker.C:
			logrus.WithFields(logrus.Fields{
				"player connected": g.LenPlayersConnectedWithLock(),
				"status":           g.gameStatus,
			}).Info("new player joined")
		default:
		}
	}
}
