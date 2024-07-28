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

	syncCompleted := make(chan struct{}, 1000)
	newServerPeer := func(peer p2p.Peer, err chan node.PeerErr) node.StartStop {
		chHeaders := make(chan *p2p.MsgHeaders, 1000)
		chBlock := make(chan *p2p.MsgBlock, 1000)
		chProcessedHeaders := make(chan struct{})
		expectedStartFromHash := make(chan [32]byte, 1000)
		outgoingMsgs := make(chan *p2p.Message, 1000)
		notifyForExpectedBlockHeaders := make(chan []p2p.BlockHeader, 1000)

		blockValidator := node.NewBlockValidator(blockRepo)
		msgHandlers := []node.StartStop{
			node.NewMsgHeaderHandler(cfg.Network, outgoingMsgs, chHeaders, expectedStartFromHash, syncCompleted, notifyForExpectedBlockHeaders),
			node.NewMsgBlockHandler(blockRepo, blockValidator, chBlock, chProcessedHeaders, notifyForExpectedBlockHeaders),
		}

		handlersManager := node.NewMessageHandlersManager(msgHandlers)

		headersRequester := sync.NewHeadersRequester(cfg.Network, blockRepo, outgoingMsgs, expectedStartFromHash)

		//processedBlocks := make(chan p2p.MsgBlock)
		peerSync := sync.NewPeerSync(headersRequester, cfg.SyncWait, chProcessedHeaders)
		nmrw := network.NewMessageReadWriter(cfg.ReadTimeout, cfg.WriteTimeout)
		//msgHeaders := make(chan *p2p.MsgHeaders)
		//msgBlocks := make(chan *p2p.MsgBlock)
		return node.NewServerPeer(cfg.Network, handlersManager, peerSync, nmrw, peer, outgoingMsgs, err, chHeaders, chBlock)
	}

	hm := p2p.NewHandshakeManager()
	peerErr := make(chan node.PeerErr, 1000)
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
