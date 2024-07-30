package node

import (
	"github.com/EmilGeorgiev/btc-node/sync"
	"log"
	"sync/atomic"

	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type MsgBlockHandler struct {
	blockRepository        sync.BlockRepository
	blockValidator         sync.BlockValidator
	stop                   chan struct{}
	blocks                 <-chan *p2p.MsgBlock
	notifyProcessedHeaders chan<- sync.RequestedHeaders
	done                   chan struct{}
	isStarted              atomic.Bool
	expectedBlockHeaders   <-chan sync.RequestedHeaders
}

func NewMsgBlockHandler(br sync.BlockRepository, bv sync.BlockValidator, blocks <-chan *p2p.MsgBlock,
	processed chan<- sync.RequestedHeaders, expBlockHeaders <-chan sync.RequestedHeaders) *MsgBlockHandler {
	return &MsgBlockHandler{
		blockRepository:        br,
		blockValidator:         bv,
		blocks:                 blocks,
		notifyProcessedHeaders: processed,

		expectedBlockHeaders: expBlockHeaders,
		stop:                 make(chan struct{}, 1000),
		done:                 make(chan struct{}, 1000),
	}
}

func (mh *MsgBlockHandler) Start() {
	if mh.isStarted.Load() {
		log.Println("MsgBlockHandler is already started.")
		return
	}
	mh.isStarted.Store(true)
	go mh.handleMsgBlock()
	log.Println("START MsgBlockHandler")
}

func (mh *MsgBlockHandler) Stop() {
	if !mh.isStarted.Load() {
		log.Println("MsgBlockHandler is not started and can't be stopped.")
		return
	}
	mh.isStarted.Store(false)
	mh.stop <- struct{}{}
	<-mh.done
	log.Println("Stop MsgBlockHandler")
}

func (mh *MsgBlockHandler) handleMsgBlock() {
	var currentBlockIndex int
	var expectedHeaders sync.RequestedHeaders
	var nextBlockHeader p2p.BlockHeader
	for {
		select {
		case <-mh.stop:
			mh.done <- struct{}{}
			return
		case expectedHeaders = <-mh.expectedBlockHeaders:
			if len(expectedHeaders.BlockHeaders) > 0 {
				nextBlockHeader = expectedHeaders.BlockHeaders[0]
				log.Printf("Set expected headers in BlockHandler.  len: %d\n", len(expectedHeaders.BlockHeaders))
				log.Printf("First block hash in exp headers is: %x\n", p2p.Reverse(Hash(expectedHeaders.BlockHeaders[0])))
				log.Printf("Prev before exp headers is: %x\n", p2p.Reverse(expectedHeaders.BlockHeaders[0].PrevBlockHash))
			}

		case block := <-mh.blocks:
			log.Printf("Handle new block: %x\n", p2p.Reverse(block.GetHash()))
			if Hash(nextBlockHeader) != block.GetHash() {
				log.Printf("unexpected block: %x\n", p2p.Reverse(block.GetHash()))
				log.Printf("Expected block is: %x\n", p2p.Reverse(Hash(nextBlockHeader)))
				continue
			}
			currentBlockIndex++
			if currentBlockIndex < len(expectedHeaders.BlockHeaders) {
				nextBlockHeader = expectedHeaders.BlockHeaders[currentBlockIndex]
			}

			log.Println("validate block")
			if err := mh.blockValidator.Validate(block); err != nil {
				log.Printf("block is not valid: %s ", err)
				continue
			}

			log.Printf("save block: %x\n", p2p.Reverse(block.GetHash()))
			if err := mh.blockRepository.Save(*block); err != nil {
				log.Println("failed to save block: ", err)
				continue
			}

			if currentBlockIndex >= len(expectedHeaders.BlockHeaders) {
				log.Printf("current block index: %d is >= len(expectedHeaders): %d\n", currentBlockIndex, len(expectedHeaders.BlockHeaders))
				log.Println("Notify PeerSync to send new requests headers")
				currentBlockIndex = 0
				mh.notifyProcessedHeaders <- expectedHeaders
				expectedHeaders = sync.RequestedHeaders{}

			}
		}
	}
}
