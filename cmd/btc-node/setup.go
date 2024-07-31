package main

import (
	"bytes"
	"encoding/hex"
	"github.com/EmilGeorgiev/btc-node/db"
	"github.com/EmilGeorgiev/btc-node/network"
	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/EmilGeorgiev/btc-node/sync"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/EmilGeorgiev/btc-node/node"
)

var genesysBlock = "0100000000000000000000000000000000000000000000000000000000000000000000003ba3edfd7a7b12b27ac72c3e67768f617fc81bc3888a51323a9fb8aa4b1e5e4a29ab5f49ffff001d1dac2b7c0101000000010000000000000000000000000000000000000000000000000000000000000000ffffffff4d04ffff001d0104455468652054696d65732030332f4a616e2f32303039204368616e63656c6c6f72206f6e206272696e6b206f66207365636f6e64206261696c6f757420666f722062616e6b73ffffffff0100f2052a01000000434104678afdb0fe5548271967f1a67130b7105cd6a828e03909a67962e0ea1f61deb649f6bc3f4cef38c4f35504e51ec112de5c384df7ba0b8d578a4c702b6bf11d5fac00000000"

func Run(cfg Config) {

	boltDB, err := db.NewBoltDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("can't initialize BoltDB: %s", err)
	}
	blockRepo, err := db.NewBlockRepo(boltDB.DB)
	if err != nil {
		log.Fatalf("can't initialize Block repository: %s", err)
	}

	//storeGenesysBlock(blockRepo)

	syncCompleted := make(chan struct{}, 1000)
	//                        notify the block hash from which headers will come
	//PeersChaub.getOverview  --------------------------------------------------->  Headershandler
	//
	//
	//                           notify for validated and proccesed
	//PeerSync.getChainOverview  <------------------------------------------------- HeadersHandler
	//
	//                 send what Blocks with Headers to be expected
	//Headershandler   ------------------------------------------------------------> BlockHandler
	//
	//                notify for processed and saved the last block header
	//PeerSync.Sync  <------------------------------------------------------------- BlockHandler
	//

	requestHeaders := make(chan sync.RequestedHeaders, 1000)
	newServerPeer := func(peer p2p.Peer, err chan node.PeerErr) node.PeerConnectionManager {
		chHeaders := make(chan *p2p.MsgHeaders, 1000)
		chBlock := make(chan *p2p.MsgBlock, 1000)
		//chProcessedHeaders := make(chan struct{})
		expectedStartFromHash := make(chan [32]byte, 1000)
		outgoingMsgs := make(chan *p2p.Message, 1000)
		//notifyForExpectedBlockHeaders := make(chan []p2p.BlockHeader, 1000)

		blockValidator := node.NewBlockValidator(blockRepo)
		msgHandlers := []node.StartStop{
			node.NewMsgHeaderHandler(cfg.Network, outgoingMsgs, chHeaders, expectedStartFromHash, syncCompleted, requestHeaders),
			node.NewMsgBlockHandler(blockRepo, blockValidator, chBlock, requestHeaders, requestHeaders),
		}
		overViewMsgHandlers := msgHandlers[:1]
		handlersManager := node.NewMessageHandlersManager(msgHandlers, overViewMsgHandlers)

		headersRequester := sync.NewHeadersRequester(cfg.Network, blockRepo, outgoingMsgs, expectedStartFromHash)

		//processedBlocks := make(chan p2p.MsgBlock)
		peerSync := sync.NewPeerSync(headersRequester, cfg.SyncWait, requestHeaders)
		nmrw := network.NewMessageReadWriter(cfg.ReadTimeout, cfg.WriteTimeout)
		//msgHeaders := make(chan *p2p.MsgHeaders)
		//msgBlocks := make(chan *p2p.MsgBlock)
		return node.NewServerPeer(cfg.Network, handlersManager, peerSync, nmrw, peer, outgoingMsgs, err, chHeaders, chBlock)
	}

	hm := p2p.NewHandshakeManager()
	peerErr := make(chan node.PeerErr, 1000)
	n, err := node.New(cfg.Network, cfg.UserAgent, newServerPeer, cfg.PeerAddrs, peerErr, syncCompleted, hm, cfg.GetNextPeerConnMngWait, cfg.ReconnectWait)
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

	log.Println("Server stopped gracefully.")
}

func storeGenesysBlock(blockRepo sync.BlockRepository) {
	b, _ := hex.DecodeString(genesysBlock)
	buf := bytes.NewBuffer(b)
	var msg p2p.Message

	if err := binary.NewDecoder(buf).Decode(&msg); err != nil {
		panic(err)
	}

	buf = bytes.NewBuffer(msg.Payload)
	var msgBlock p2p.MsgBlock
	if err := binary.NewDecoder(buf).Decode(&msgBlock); err != nil {
		panic(err)
	}

	if err := blockRepo.Save(msgBlock); err != nil {
		panic(err)
	}
}
