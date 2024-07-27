package main

import (
	"fmt"
	"github.com/EmilGeorgiev/btc-node/db"
	"github.com/EmilGeorgiev/btc-node/network"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/EmilGeorgiev/btc-node/sync"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/EmilGeorgiev/btc-node/node"
)

func Run(cfg Config) {

	boltDB, err := db.NewBoltDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("can't initialize BoltDB: %s", err)
	}
	blockRepo, err := db.NewBlockRepo(boltDB.DB)
	if err != nil {
		log.Fatalf("can't initialize Block repository: %s", err)
	}

	syncCompleted := make(chan struct{})
	newServerPeer := func(peer p2p.Peer, err chan node.PeerErr) node.PeerConnectionManager {
		chHeaders := make(chan *p2p.MsgHeaders)
		chBlock := make(chan *p2p.MsgBlock)
		chProcessedBlock := make(chan *p2p.MsgBlock)
		expectedHeaders := make(chan [32]byte)
		outgoingMsgs := make(chan *p2p.Message)

		blockValidator := node.NewBlockValidator()
		msgHandlers := []node.StartStop{
			node.NewMsgHeaderHandler(cfg.Network, outgoingMsgs, chHeaders, expectedHeaders, syncCompleted),
			node.NewMsgBlockHandler(blockRepo, blockValidator, chBlock, chProcessedBlock),
		}

		handlersManager := node.NewMessageHandlersManager(msgHandlers)

		headersRequester := sync.NewHeadersRequester(cfg.Network, blockRepo, outgoingMsgs, expectedHeaders)

		processedBlocks := make(chan p2p.MsgBlock)
		peerSync := sync.NewPeerSync(headersRequester, cfg.SyncWait, processedBlocks)
		nmrw := network.NewMessageReadWriter(cfg.ReadTimeout, cfg.WriteTimeout)
		//msgHeaders := make(chan *p2p.MsgHeaders)
		//msgBlocks := make(chan *p2p.MsgBlock)
		return node.NewServerPeer(cfg.Network, handlersManager, peerSync, nmrw, peer, outgoingMsgs, err, chHeaders, chBlock)
	}

	hm := p2p.NewHandshakeManager()
	peerErr := make(chan node.PeerErr)
	n, err := node.New(cfg.Network, cfg.UserAgent, newServerPeer, cfg.PeerAddrs, peerErr, syncCompleted, hm, cfg.GetNextPeerConnMngWait)
	if err != nil {
		log.Fatalf("failed to initialize the Node: %s", err)
	}

	n.Start()

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
