package main

import (
	"time"

	"github.com/k0yote/k0porker/p2p"
)

func makeServerAndStart(addr, apiAddr string) *p2p.Server {
	cfg := p2p.ServerConfig{
		Version:       "0.1-alpha",
		ListenAddr:    addr,
		APIListneAddr: apiAddr,
		GameVariant:   p2p.TexasHoldem,
	}

	server := p2p.NewServer(cfg)

	go server.Start()

	time.Sleep(1 * time.Second)

	return server
}

func main() {

	playerA := makeServerAndStart(":3000", ":3001")
	playerB := makeServerAndStart(":4000", ":4001")
	playerC := makeServerAndStart(":5000", ":5001")
	playerD := makeServerAndStart(":6000", ":6001")
	// playerE := makeServerAndStart(":7000")
	// playerF := makeServerAndStart(":8000")

	time.Sleep(200 * time.Millisecond)
	playerB.Connect(playerA.ListenAddr)
	time.Sleep(200 * time.Millisecond)
	playerC.Connect(playerB.ListenAddr)
	time.Sleep(200 * time.Millisecond)
	playerD.Connect(playerC.ListenAddr)
	// time.Sleep(200 * time.Millisecond)
	// playerE.Connect(playerD.ListenAddr)
	// time.Sleep(200 * time.Millisecond)
	// playerF.Connect(playerE.ListenAddr)

	select {}
}
