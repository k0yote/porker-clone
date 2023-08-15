package main

import (
	"log"
	"time"

	"github.com/k0yote/k0porker/p2p"
)

func main() {

	// d := deck.New()
	// fmt.Println(d)

	server := p2p.NewServer(p2p.ServerConfig{
		Version:     "0.1-alpha",
		ListenAddr:  ":3000",
		GameVariant: p2p.TexasHoldem,
	})

	go server.Start()

	time.Sleep(1 * time.Second)

	remoteServer := p2p.NewServer(p2p.ServerConfig{
		Version:     "0.1-alpha",
		ListenAddr:  ":4000",
		GameVariant: p2p.TexasHoldem,
	})
	go remoteServer.Start()
	if err := remoteServer.Connect(":3000"); err != nil {
		log.Fatalln(err)
	}

	select {}
}
