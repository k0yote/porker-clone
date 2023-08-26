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

type PlayersList struct {
	lock sync.RWMutex
	list []string
}

func NewPlayersList() *PlayersList {
	return &PlayersList{
		list: []string{},
	}
}

func (p *PlayersList) List() []string {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return p.list
}

func (p *PlayersList) len() int {
	p.lock.RLock()
	defer p.lock.RUnlock()

	return len(p.list)
}

func (p *PlayersList) add(addr string) {
	p.lock.Lock()
	defer p.lock.Unlock()

	p.list = append(p.list, addr)
	sort.Sort(p)
}

func (p *PlayersList) getIndex(addr string) int {
	p.lock.RLock()
	defer p.lock.RUnlock()

	for i := 0; i < len(p.list); i++ {
		if addr == p.list[i] {
			return i
		}
	}

	return -1
}

func (p *PlayersList) get(index any) string {
	p.lock.RLock()
	defer p.lock.RUnlock()

	var i int
	switch v := index.(type) {
	case int:
		i = v
	case int32:
		i = int(v)
	}

	if len(p.list) < i {
		panic("the given index is out of bounds")
	}

	return p.list[i]
}

func (p *PlayersList) Len() int { return len(p.list) }
func (p *PlayersList) Swap(i, j int) {
	p.list[i], p.list[j] = p.list[j], p.list[i]
}
func (p *PlayersList) Less(i, j int) bool {
	portI, _ := strconv.Atoi(p.list[i][1:])
	portJ, _ := strconv.Atoi(p.list[j][1:])

	return portI < portJ
}

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

func (pa *PlayerActionRecv) clear() {
	pa.mu.Lock()
	defer pa.mu.Unlock()

	pa.recvActions = make(map[string]MessagePlayerAction)
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

	playersList *PlayersList

	// playersReadyList *PlayersList

	table *Table
}

func NewGame(addr string, broadcastCh chan BroadcastTo) *Game {
	g := &Game{
		listenAddr:    addr,
		broadcastCh:   broadcastCh,
		currentStatus: NewAtomicInt(int32(GameStatusConnected)),
		playersReady:  NewPlayersReady(),
		// playersReadyList:    NewPlayersList(),
		playersList:         NewPlayersList(),
		recvPlayerActions:   NewPlayerActionRecv(),
		currentPlayerAction: NewAtomicInt(0),
		currentPlayerTurn:   NewAtomicInt(0),
		currentDealer:       NewAtomicInt(0),
		table:               NewTable(6),
	}

	g.playersList.add(addr)

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
	currentPlayerAddr := g.playersList.get(g.currentPlayerTurn.Get())

	return currentPlayerAddr == from
}

func (g *Game) isFromCurrentDealer(from string) bool {
	return g.playersList.get(g.currentDealer.Get()) == from
}

func (g *Game) handlePlayerAction(from string, action MessagePlayerAction) error {
	if !g.canTakeAction(from) {
		return fmt.Errorf("player (%s) taking action before his turn", from)
	}

	if action.CurrentGameStatus != GameStatus(g.currentStatus.Get()) && !g.isFromCurrentDealer(from) {
		return fmt.Errorf("player (%s) has not the correct game status (%s)", from, action.CurrentGameStatus)
	}

	g.recvPlayerActions.addAction(from, action)

	if g.playersList.get(g.currentDealer.Get()) == from {
		g.advanceToNextRound()
	}

	g.incNextPlayer()

	logrus.WithFields(logrus.Fields{
		"we":     g.listenAddr,
		"from":   from,
		"action": action,
	}).Info("recv player action")

	return nil
}

func (g *Game) TakeAction(action PlayerAction, value int) error {
	if !g.canTakeAction(g.listenAddr) {
		return fmt.Errorf("taking action before its my turn %s", g.listenAddr)
	}

	g.currentPlayerAction.Set((int32)(action))

	g.incNextPlayer()

	if g.listenAddr == g.playersList.get(g.currentDealer.Get()) {
		g.advanceToNextRound()
	}

	a := MessagePlayerAction{
		CurrentGameStatus: GameStatus(g.currentStatus.Get()),
		Action:            action,
		Value:             value,
	}

	g.sendToPlayers(a, g.getOtherPlayers()...)

	return nil
}

func (g *Game) advanceToNextRound() {
	if GameStatus(g.currentStatus.Get()) == GameStatusRiver {
		panic("")
	}

	g.recvPlayerActions.clear()
	g.currentPlayerAction.Set(int32(PlayerActionNone))
	g.currentStatus.Set(int32(g.getNextGameStatus()))
}

func (g *Game) incNextPlayer() {
	if g.playersList.len()-1 == int(g.currentPlayerTurn.Get()) {
		g.currentPlayerTurn.Set(0)
		return
	}

	g.currentPlayerTurn.Inc()
}

func (g *Game) SetStatus(status GameStatus) {
	g.setStatus(status)
	g.table.SetPlayerStatus(g.listenAddr, status)
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
	currentDealerAddr := g.playersList.get(g.currentDealer.Get())

	return currentDealerAddr, g.listenAddr == currentDealerAddr
}

func (g *Game) ShuffleAndEncrypt(from string, deck [][]byte) error {
	prevPlayer, err := g.table.GetPlayerBefore(g.listenAddr)
	if err != nil {
		panic(err)
	}
	if from != prevPlayer.addr {
		return fmt.Errorf("received encrypted deck from the wrong player (%s) should be (%s)", from, prevPlayer.addr)
	}

	_, isDealer := g.getCurrentDealerAddr()

	if isDealer && from == prevPlayer.addr {
		g.setStatus(GameStatusPreFlop)
		g.table.SetPlayerStatus(g.listenAddr, GameStatusPreFlop)
		g.sendToPlayers(MessagePreFlop{}, g.getOtherPlayers()...)
		return nil
	}

	dealToPlayer, err := g.table.GetPlayerAfter(g.listenAddr)
	if err != nil {
		panic(err)
	}
	logrus.WithFields(logrus.Fields{
		"recvFromPlayer": from,
		"we":             g.listenAddr,
		"dealToPlayer":   dealToPlayer.addr,
	}).Info("received cards and going to shuffle")

	g.sendToPlayers(MessageEncDeck{Deck: [][]byte{}}, dealToPlayer.addr)
	g.setStatus(GameStatusDealing)

	return nil
}

func (g *Game) InitiateShuffleAndDeal() {
	fmt.Println("================================")
	fmt.Println(g.listenAddr)
	fmt.Println("================================")

	// dealToPlayerAddr := g.playersList.get(g.getNextPositionOnTable())
	dealToPlayer, err := g.table.GetPlayerAfter(g.listenAddr)
	if err != nil {
		panic(err)
	}
	g.setStatus(GameStatusDealing)
	g.sendToPlayers(MessageEncDeck{Deck: [][]byte{}}, dealToPlayer.addr)

	logrus.WithFields(logrus.Fields{
		"we": g.listenAddr,
		"to": dealToPlayer.addr,
	}).Info("dealing cards")
}

func (g *Game) maybeDeal() {
	if GameStatus(g.currentStatus.Get()) == GameStatusPlayerReady {
		g.InitiateShuffleAndDeal()
	}
}

func (g *Game) SetPlayerReady(addr string) {
	tablePos := g.playersList.getIndex(addr)
	g.table.AddPlayerOnPosition(addr, tablePos)

	g.playersReady.addRecvStatus(addr)

	if g.table.LenPlayers() < 2 {
		return
	}

	if _, areWeDealer := g.getCurrentDealerAddr(); areWeDealer {
		go func() {
			time.Sleep(5 * time.Second)
			g.maybeDeal()
		}()
	}
}

func (g *Game) SetReady() {
	tablePos := g.playersList.getIndex(g.listenAddr)
	g.table.AddPlayerOnPosition(g.listenAddr, tablePos)
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
	}) //.Info("sending payload to player")

}

func (g *Game) AddPlayer(from string) {
	g.playersList.add(from)
}

func (g *Game) loop() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		<-ticker.C

		currentDealer, _ := g.getCurrentDealerAddr()
		logrus.WithFields(logrus.Fields{
			"we":             g.listenAddr,
			"playerList":     g.playersList.List(),
			"gameState":      GameStatus(g.currentStatus.Get()),
			"currentDealer":  currentDealer,
			"nextPlayerTurn": g.currentPlayerTurn,
			// "table":          g.table,
		}).Info("new player joined")
	}
}

func (g *Game) getOtherPlayers() []string {
	players := []string{}

	for _, addr := range g.playersList.List() {
		if addr == g.listenAddr {
			continue
		}

		players = append(players, addr)
	}

	return players
}

func (g *Game) getPositionOnTable() int {
	for i, addr := range g.playersList.List() {
		if g.listenAddr == addr {
			return i
		}
	}

	panic("player does not exist in the playersList; that should not happen")
}
