package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/EmilGeorgiev/btc-node/node"
)

func Run(cfg Config) {

	n, err := node.New(cfg.Network, cfg.UserAgent, cfg.ReadTimeout)
	if err != nil {
		log.Fatalf("failed to initialize the Node: %s", err)
	}

	n.ConnectToPeers(cfg.PeerAddrs)
	//if the node is run for the first time or was down for some time it can miss some of the new
	// blocks and information. So it should sync with other peers
	n.Sync()

	// Create a signal channel to listen for interrupt or termination signals
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	// Wait for a signal
	s := <-signalChan
	log.Printf("received signal: %s. Stoping the node.\n", s.String())

	// Stop the node gracefully
	n.Stop()

	fmt.Println("Server stopped gracefully.")
}
