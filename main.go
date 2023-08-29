package main

import (
	"net/http"
	"time"

	"github.com/k0yote/k0porker/p2p"
)

func makeServerAndStart(addr, apiAddr string) *p2p.Node {
	cfg := p2p.ServerConfig{
		Version:       "0.1-alpha",
		ListenAddr:    addr,
		APIListneAddr: apiAddr,
		GameVariant:   p2p.TexasHoldem,
	}

	server := p2p.NewNode(cfg)

	go server.Start()

	time.Sleep(200 * time.Millisecond)

	return server
}

func main() {
	node1 := makeServerAndStart(":3000", ":3001")
	node2 := makeServerAndStart(":4000", ":4001")
	node3 := makeServerAndStart(":5000", ":5001")
	node4 := makeServerAndStart(":6000", ":6001")

	node2.Connect(node1.ListenAddr)
	node3.Connect(node2.ListenAddr)
	node4.Connect(node3.ListenAddr)

	go func() {
		requestAction()
	}()

	select {}
}

func requestAction() {
	reqList := []string{
		"http://localhost:3001/takeseat",
		"http://localhost:4001/takeseat",
		"http://localhost:5001/takeseat",
		"http://localhost:6001/takeseat",

		// "http://localhost:4001/fold",
		// "http://localhost:5001/fold",
		// "http://localhost:6001/fold",
		// "http://localhost:3001/fold",

		// "http://localhost:4001/fold",
		// "http://localhost:5001/fold",
		// "http://localhost:6001/fold",
		// "http://localhost:3001/fold",

		// "http://localhost:4001/fold",
		// "http://localhost:5001/fold",
		// "http://localhost:6001/fold",
		// "http://localhost:3001/fold",

		// "http://localhost:4001/fold",
		// "http://localhost:5001/fold",
		// "http://localhost:6001/fold",
		// "http://localhost:3001/fold",
	}

	for _, action := range reqList {
		time.Sleep(2 * time.Second)
		http.Get(action)
	}
}
