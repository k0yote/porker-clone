package p2p

import (
	"fmt"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

type AtomicInt struct {
	value int32
}

func NewAtomicInt(value int32) *AtomicInt {
	return &AtomicInt{
		value: value,
	}
}

func (a *AtomicInt) Set(value int32) {
	atomic.StoreInt32(&a.value, value)
}

func (a *AtomicInt) Get() int32 {
	return atomic.LoadInt32(&a.value)
}

func (a *AtomicInt) Inc() {
	currentValue := a.Get()
	a.Set(currentValue + 1)
}

func (a *AtomicInt) String() string {
	return fmt.Sprintf("%d", a.value)
}

type PlayerActionRecv struct {
	mu          sync.RWMutex
	recvActions map[string]MessagePlayerAction
}

func NewPlayerActionRecv() *PlayerActionRecv {
	return &PlayerActionRecv{
		recvActions: make(map[string]MessagePlayerAction),
	}
}

func (pa *PlayerActionRecv) addAction(from string, action MessagePlayerAction) {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	pa.recvActions[from] = action
}

type PlayersReady struct {
	mu         sync.RWMutex
	recvStatus map[string]bool
}

func NewPlayersReady() *PlayersReady {
	return &PlayersReady{
		recvStatus: make(map[string]bool),
	}
}

func (pr *PlayersReady) addRecvStatus(from string) {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.recvStatus[from] = true
}

func (pr *PlayersReady) len() int {
	pr.mu.RLock()
	defer pr.mu.RUnlock()

	return len(pr.recvStatus)
}

func (pr *PlayersReady) clear() {
	pr.mu.Lock()
	defer pr.mu.Unlock()

	pr.recvStatus = make(map[string]bool)
}

type Game struct {
	listenAddr  string
	broadcastCh chan BroadcastTo

	currentPlayerAction *AtomicInt

	currentStatus *AtomicInt

	currentDealer     *AtomicInt
	currentPlayerTurn *AtomicInt

	playersReady      *PlayersReady
	recvPlayerActions *PlayerActionRecv

	playersList PlayersList
}

func NewGame(addr string, broadcastCh chan BroadcastTo) *Game {
	g := &Game{
		listenAddr:          addr,
		broadcastCh:         broadcastCh,
		currentStatus:       NewAtomicInt(int32(GameStatusConnected)),
		playersReady:        NewPlayersReady(),
		playersList:         PlayersList{},
		recvPlayerActions:   NewPlayerActionRecv(),
		currentPlayerAction: NewAtomicInt(0),
		currentPlayerTurn:   NewAtomicInt(0),
		currentDealer:       NewAtomicInt(0),
	}

	g.playersList = append(g.playersList, addr)

	go g.loop()

	return g
}

func (g *Game) getNextGameStatus() GameStatus {
	switch GameStatus(g.currentStatus.Get()) {
	case GameStatusPreFlop:
		return GameStatusFlop
	case GameStatusFlop:
		return GameStatusTurn
	case GameStatusTurn:
		return GameStatusRiver
	default:
		panic("invalid game status")
	}
}

func (g *Game) canTakeAction(from string) bool {
	// currentPlayerAddr := g.playersList[g.currentPlayerTurn.Get()]
	currentPlayerAddr := g.playersList[g.currentPlayerTurn.Get()]

	return currentPlayerAddr == from
}

func (g *Game) handlePlayerAction(from string, action MessagePlayerAction) error {
	if !g.canTakeAction(from) {
		return fmt.Errorf("player (%s) taking action before his turn", from)
	}

	if action.CurrentGameStatus != GameStatus(g.currentStatus.Get()) {
		return fmt.Errorf("player (%s) has not the correct game status (%s)", from, action.CurrentGameStatus)
	}

	logrus.WithFields(logrus.Fields{
		"we":     g.listenAddr,
		"from":   from,
		"action": action,
	}).Info("recv player action")

	if g.playersList[g.currentDealer.Get()] == from {
		g.currentStatus.Set(int32(g.getNextGameStatus()))
	}

	g.recvPlayerActions.addAction(from, action)

	g.incNextPlayer()

	return nil
}

func (g *Game) TakeAction(action PlayerAction, value int) error {
	if !g.canTakeAction(g.listenAddr) {
		return fmt.Errorf("i am taking action before its my turn %s", g.listenAddr)
	}

	g.currentPlayerAction.Set((int32)(action))
	// if action == PlayerActionFold {

	// }

	// if action == PlayerActionCheck {

	// }

	g.incNextPlayer()

	if g.listenAddr == g.playersList[g.currentDealer.Get()] {
		g.currentStatus.Set(int32(g.getNextGameStatus()))
	}

	a := MessagePlayerAction{
		CurrentGameStatus: GameStatus(g.currentStatus.Get()),
		Action:            action,
		Value:             value,
	}

	g.sendToPlayers(a, g.getOtherPlayers()...)

	return nil
}

func (g *Game) incNextPlayer() {
	if len(g.playersList)-1 == int(g.currentPlayerTurn.Get()) {
		g.currentPlayerTurn.Set(0)
		return
	}

	g.currentPlayerTurn.Inc()
}

func (g *Game) SetStatus(status GameStatus) {
	g.setStatus(status)
}

func (g *Game) setStatus(s GameStatus) {
	if s == GameStatusPreFlop {
		g.incNextPlayer()
	}

	if GameStatus(g.currentStatus.Get()) != s {
		g.currentStatus.Set(int32(s))
	}
}

func (g *Game) getCurrentDealerAddr() (string, bool) {
	currentDealerAddr := g.playersList[g.currentDealer.Get()]

	return currentDealerAddr, g.listenAddr == currentDealerAddr
}

func (g *Game) SetPlayerReady(from string) {
	logrus.WithFields(logrus.Fields{
		"websocket": g.listenAddr,
		"player":    from,
	}).Info("setting player status to ready")

	g.playersReady.addRecvStatus(from)

	if g.playersReady.len() < 2 {
		return
	}

	//g.playersReady.clear()

	if _, ok := g.getCurrentDealerAddr(); ok {
		// fmt.Println("round can be started we hanve players: ", g.playersReady.len())
		// fmt.Println("we are the dealer: ", g.listenAddr)
		g.InitiateShuffleAndDeal()
	}

}

func (g *Game) ShuffleAndEncrypt(from string, deck [][]byte) error {
	prevPlayerAddr := g.playersList[g.getPrevPositionOnTable()]
	if from != prevPlayerAddr {
		return fmt.Errorf("received encrypted deck from the wrong player (%s) should be (%s)", from, prevPlayerAddr)
	}

	_, isDealer := g.getCurrentDealerAddr()

	if isDealer && from == prevPlayerAddr {
		g.setStatus(GameStatusPreFlop)
		g.sendToPlayers(MessagePreFlop{}, g.getOtherPlayers()...)
		return nil
	}

	dealToPlayer := g.playersList[g.getNextPositionOnTable()]

	logrus.WithFields(logrus.Fields{
		"recvFromPlayer": from,
		"we":             g.listenAddr,
		"dealToPlayer":   dealToPlayer,
	}).Info("received cards and going to shuffle")

	g.sendToPlayers(MessageEncDeck{Deck: [][]byte{}}, dealToPlayer)
	g.setStatus(GameStatusDealing)

	return nil
}

func (g *Game) InitiateShuffleAndDeal() {
	dealToPlayerAddr := g.playersList[g.getNextPositionOnTable()]
	g.setStatus(GameStatusDealing)
	g.sendToPlayers(MessageEncDeck{Deck: [][]byte{}}, dealToPlayerAddr)

	logrus.WithFields(logrus.Fields{
		"we": g.listenAddr,
		"to": dealToPlayerAddr,
	}).Info("dealing cards")
}

func (g *Game) SetReady() {
	g.playersReady.addRecvStatus(g.listenAddr)
	g.sendToPlayers(MessageReady{}, g.getOtherPlayers()...)
	g.setStatus(GameStatusPlayerReady)
}

func (g *Game) sendToPlayers(payload any, addr ...string) {
	g.broadcastCh <- BroadcastTo{
		To:      addr,
		Payload: payload,
	}

	logrus.WithFields(logrus.Fields{
		"payload": payload,
		"player":  addr,
		"we":      g.listenAddr,
	}).Info("sending payload to player")

}

func (g *Game) AddPlayer(from string) {
	g.playersList = append(g.playersList, from)
	sort.Sort(g.playersList)

	// g.playersReady.addRecvStatus(from)
}

func (g *Game) loop() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		<-ticker.C

		currentDealer, _ := g.getCurrentDealerAddr()
		logrus.WithFields(logrus.Fields{
			"action":              PlayerActionFold,
			"players":             g.playersList,
			"gameStatus":          GameStatus(g.currentStatus.Get()),
			"we":                  g.listenAddr,
			"currentDealer":       currentDealer,
			"nextPlayerTurn":      g.currentPlayerTurn,
			"playerActions":       g.recvPlayerActions.recvActions,
			"currentPlayerAction": PlayerAction(g.currentPlayerAction.Get()),
		}).Info("new player joined")
	}
}

func (g *Game) getOtherPlayers() []string {
	players := []string{}

	for _, addr := range g.playersList {
		if addr == g.listenAddr {
			continue
		}

		players = append(players, addr)
	}

	return players
}

func (g *Game) getPrevPositionOnTable() int {
	ourPosition := g.getPositionOnTable()

	if ourPosition == 0 {
		return len(g.playersList) - 1
	}

	return ourPosition - 1
}

func (g *Game) getPositionOnTable() int {
	for i, addr := range g.playersList {
		if g.listenAddr == addr {
			return i
		}
	}

	panic("player does not exist in the playersList; that should not happen")
}

func (g *Game) getNextPositionOnTable() int {
	ourPosition := g.getPositionOnTable()
	if ourPosition == len(g.playersList)-1 {
		return 0
	}

	return ourPosition + 1
}

type PlayersList []string

func (list PlayersList) Len() int      { return len(list) }
func (list PlayersList) Swap(i, j int) { list[i], list[j] = list[j], list[i] }
func (list PlayersList) Less(i, j int) bool {
	portI, _ := strconv.Atoi(list[i][1:])
	portJ, _ := strconv.Atoi(list[j][1:])

	return portI < portJ
}
