package p2p

import (
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

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

	currentStatus GameStatus

	playersReady *PlayersReady
	playersList  PlayersList
}

func NewGame(addr string, broadcastCh chan BroadcastTo) *Game {
	g := &Game{
		listenAddr:    addr,
		broadcastCh:   broadcastCh,
		currentStatus: GameStatusConnected,
		playersReady:  NewPlayersReady(),
		playersList:   PlayersList{},
	}

	go g.loop()

	return g
}

func (g *Game) setStatus(s GameStatus) {
	if g.currentStatus != s {
		atomic.StoreInt32((*int32)(&g.currentStatus), (int32)(s))
	}
}

func (g *Game) SetReady() {
	g.playersReady.addRecvStatus(g.listenAddr)
	g.SendToPlayer(MessageReady{}, g.getOtherPlayers()...)
	g.setStatus(GameStatusPlayerReady)
}

func (g *Game) SendToPlayer(payload any, addr ...string) {
	g.broadcastCh <- BroadcastTo{
		To:      addr,
		Payload: payload,
	}

	logrus.WithFields(logrus.Fields{
		"payload": payload,
		"player":  addr,
	}).Info("sending payload to player")

}

func (g *Game) AddPlayer(from string) {
	g.playersList = append(g.playersList, from)
	sort.Sort(g.playersList)

	g.playersReady.addRecvStatus(from)
}

func (g *Game) loop() {
	ticker := time.NewTicker(5 * time.Second)
	for {
		<-ticker.C
		logrus.WithFields(logrus.Fields{
			"players": g.playersList,
			"status":  g.currentStatus,
			"we":      g.listenAddr,
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

// type GameState struct {
// 	listenAddr  string
// 	broadcastCh chan BroadcastTo
// 	isDealer    bool
// 	gameStatus  GameStatus

// 	playersList PlayersList
// 	playersLock sync.RWMutex
// 	players     map[string]*Player
// }

// func NewGameState(addr string, broadcastCh chan BroadcastTo) *GameState {
// 	g := &GameState{
// 		listenAddr:  addr,
// 		broadcastCh: broadcastCh,
// 		isDealer:    false,
// 		gameStatus:  GameStatusWaitingForCards,
// 		players:     make(map[string]*Player),
// 	}

// 	g.AddPlayer(addr, GameStatusWaitingForCards)

// 	go g.loop()

// 	return g
// }

// func (g *GameState) SetStatus(s GameStatus) {
// 	if g.gameStatus != s {
// 		atomic.StoreInt32((*int32)(&g.gameStatus), (int32)(s))
// 		g.SetPlayerStatus(g.listenAddr, s)
// 	}
// }

// func (g *GameState) playersWaitingForCards() int {
// 	totalPlayers := 0

// 	for i := 0; i < len(g.playersList); i++ {
// 		if g.playersList[i].Status == GameStatusWaitingForCards {
// 			totalPlayers++
// 		}
// 	}
// 	return totalPlayers
// }

// func (g *GameState) CheckNeedDealCards() {
// 	playersWaiting := g.playersWaitingForCards()

// 	if playersWaiting == len(g.players) &&
// 		g.isDealer &&
// 		g.gameStatus == GameStatusWaitingForCards {
// 		logrus.WithFields(logrus.Fields{
// 			"addr": g.listenAddr,
// 		}).Info("need to deal cards")

// 		g.InitiateShuffleAndDeal()
// 	}
// }

// func (g *GameState) GetPlayersWithStatus(s GameStatus) []string {
// 	players := []string{}
// 	for addr, player := range g.players {
// 		if player.Status == s {
// 			players = append(players, addr)
// 		}
// 	}

// 	return players
// }

// func (g *GameState) getPrevPositionOnTable() int {
// 	ourPosition := g.getPositionOnTable()

// 	if ourPosition == 0 {
// 		return len(g.playersList) - 1
// 	}

// 	return ourPosition - 1
// }

// func (g *GameState) getPositionOnTable() int {
// 	for i, player := range g.playersList {
// 		if g.listenAddr == player.ListenAddr {
// 			return i
// 		}
// 	}

// 	panic("player does not exist in the playersList; that should not happen")
// }

// func (g *GameState) getNextPositionOnTable() int {
// 	ourPosition := g.getPositionOnTable()
// 	if ourPosition == len(g.playersList)-1 {
// 		return 0
// 	}

// 	return ourPosition + 1
// }

// func (g *GameState) ShuffleAndEncrypt(from string, deck [][]byte) error {
// 	g.SetPlayerStatus(from, GameStatusShuffleAndDeal)

// 	prevPlayer := g.playersList[g.getPrevPositionOnTable()]
// 	if g.isDealer && from == prevPlayer.ListenAddr {
// 		logrus.Info("shuffle roundtrip completed")
// 		return nil
// 	}

// 	dealToPlayer := g.playersList[g.getNextPositionOnTable()]
// 	logrus.WithFields(logrus.Fields{
// 		"recvFromPlayer": from,
// 		"we":             g.listenAddr,
// 		"dealToPlayer":   dealToPlayer,
// 	}).Info("received cards and going to shuffle")

// 	g.SendToPlayer(dealToPlayer.ListenAddr, MessageEncDeck{Deck: [][]byte{}})
// 	g.SetStatus(GameStatusShuffleAndDeal)

// 	return nil
// }

// func (g *GameState) InitiateShuffleAndDeal() {
// 	dealToPlayer := g.playersList[g.getNextPositionOnTable()]

// 	g.SetStatus(GameStatusShuffleAndDeal)
// 	g.SendToPlayer(dealToPlayer.ListenAddr, MessageEncDeck{Deck: [][]byte{}})
// }

// func (g *GameState) SendToPlayer(addr string, payload any) {
// 	g.broadcastCh <- BroadcastTo{
// 		To:      []string{addr},
// 		Payload: payload,
// 	}

// 	logrus.WithFields(logrus.Fields{
// 		"payload": payload,
// 		"player":  addr,
// 	}).Info("sending payload to player")

// }

// func (g *GameState) SendToPlayersWithStatus(payload any, s GameStatus) {
// 	players := g.GetPlayersWithStatus(s)

// 	g.broadcastCh <- BroadcastTo{
// 		To:      players,
// 		Payload: payload,
// 	}

// 	logrus.WithFields(logrus.Fields{
// 		"payload": payload,
// 		"players": players,
// 	}).Info("sending to players")
// }

// func (g *GameState) SetPlayerStatus(addr string, status GameStatus) {
// 	player, ok := g.players[addr]
// 	if !ok {
// 		panic("player not found although it should exist")
// 	}

// 	player.Status = status

// 	g.CheckNeedDealCards()
// }

// func (g *GameState) AddPlayer(addr string, status GameStatus) {
// 	g.playersLock.Lock()
// 	defer g.playersLock.Unlock()

// 	player := &Player{
// 		ListenAddr: addr,
// 	}

// 	g.players[addr] = player
// 	g.playersList = append(g.playersList, player)
// 	sort.Sort(g.playersList)

// 	g.SetPlayerStatus(addr, status)

// 	logrus.WithFields(logrus.Fields{
// 		"addr":   addr,
// 		"status": status,
// 	}).Info("new player joined")
// }

// func (g *GameState) loop() {
// 	ticker := time.NewTicker(5 * time.Second)
// 	for {
// 		<-ticker.C
// 		logrus.WithFields(logrus.Fields{
// 			"players": g.playersList,
// 			"status":  g.gameStatus,
// 			"we":      g.listenAddr,
// 		}).Info("new player joined")
// 	}
// }

type PlayersList []string

func (list PlayersList) Len() int      { return len(list) }
func (list PlayersList) Swap(i, j int) { list[i], list[j] = list[j], list[i] }
func (list PlayersList) Less(i, j int) bool {
	portI, _ := strconv.Atoi(list[i][1:])
	portJ, _ := strconv.Atoi(list[j][1:])

	return portI < portJ
}

// type PlayersList []*Player

// func (list PlayersList) Len() int      { return len(list) }
// func (list PlayersList) Swap(i, j int) { list[i], list[j] = list[j], list[i] }
// func (list PlayersList) Less(i, j int) bool {
// 	portI, _ := strconv.Atoi(list[i].ListenAddr[1:])
// 	portJ, _ := strconv.Atoi(list[j].ListenAddr[1:])

// 	return portI < portJ
// }

// type Player struct {
// 	Status     GameStatus
// 	ListenAddr string
// }

// func (p *Player) String() string {
// 	return fmt.Sprintf("%s:%s", p.ListenAddr, p.Status)
// }
