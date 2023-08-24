package main

import (
	"net/http"
	"strings"
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
	// playerC := makeServerAndStart(":5000", ":5001")
	// playerD := makeServerAndStart(":6000", ":6001")

	go func() {
		requestAction()
	}()

	time.Sleep(200 * time.Millisecond)
	playerB.Connect(playerA.ListenAddr)

	// time.Sleep(200 * time.Millisecond)
	// playerC.Connect(playerB.ListenAddr)

	// time.Sleep(200 * time.Millisecond)
	// playerD.Connect(playerC.ListenAddr)

	select {}
}

func requestAction() {
	reqList := []string{
		"http://localhost:3001/ready",
		"http://localhost:4001/ready",
		// "http://localhost:5001/reay",
		// "http://localhost:6001/reay",
		"http://localhost:4001/fold",
		// "http://localhost:5001/fold",
		// "http://localhost:6001/fold",
		"http://localhost:3001/fold",
	}

	onetime := false

	for _, action := range reqList {
		if strings.Contains(action, "ready") {
			time.Sleep(2 * time.Second)
		}

		if strings.Contains(action, "fold") && !onetime {
			time.Sleep(5 * time.Second)
		}

		http.Get(action)
	}
}
