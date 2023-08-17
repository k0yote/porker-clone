package main

import (
	"time"

	"github.com/k0yote/k0porker/p2p"
)

func makeServerAndStart(addr string) *p2p.Server {
	cfg := p2p.ServerConfig{
		Version:     "0.1-alpha",
		ListenAddr:  addr,
		GameVariant: p2p.TexasHoldem,
	}

	server := p2p.NewServer(cfg)

	go server.Start()

	time.Sleep(1 * time.Second)

	return server
}

func main() {

	playerA := makeServerAndStart(":3000")
	playerB := makeServerAndStart(":4000")
	playerC := makeServerAndStart(":5000")
	playerD := makeServerAndStart(":6000")
	playerE := makeServerAndStart(":7000")
	playerF := makeServerAndStart(":8000")

	time.Sleep(1 * time.Second)
	playerB.Connect(playerA.ListenAddr)
	time.Sleep(1 * time.Second)
	playerC.Connect(playerB.ListenAddr)
	time.Sleep(1 * time.Second)
	playerD.Connect(playerC.ListenAddr)
	time.Sleep(1 * time.Second)
	playerE.Connect(playerD.ListenAddr)
	time.Sleep(1 * time.Second)
	playerF.Connect(playerE.ListenAddr)

	select {}
}
