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
	notifyProcessedHeaders chan<- struct{}
	done                   chan struct{}
	isStarted              atomic.Bool
	expectedBlockHeaders   <-chan []p2p.BlockHeader
}

func NewMsgBlockHandler(br sync.BlockRepository, bv sync.BlockValidator, blocks <-chan *p2p.MsgBlock,
	processed chan<- struct{}, expBlockHeaders <-chan []p2p.BlockHeader) *MsgBlockHandler {
	return &MsgBlockHandler{
		blockRepository:        br,
		blockValidator:         bv,
		blocks:                 blocks,
		notifyProcessedHeaders: processed,

		expectedBlockHeaders: expBlockHeaders,
		stop:                 make(chan struct{}),
		done:                 make(chan struct{}),
	}
}

func (mh *MsgBlockHandler) Start() {
	if mh.isStarted.Load() {
		return
	}
	log.Println("Start MsgBlock Handler.")
	mh.isStarted.Store(true)
	go mh.handleMsgBlock()
}

func (mh *MsgBlockHandler) Stop() {
	if !mh.isStarted.Load() {
		return
	}
	log.Println("Stop MsgBlock handler.")
	mh.isStarted.Store(false)
	mh.stop <- struct{}{}
	<-mh.done
}

func (mh *MsgBlockHandler) handleMsgBlock() {
	var currentBlockIndex int
	var expectedHeaders []p2p.BlockHeader
	for {
		select {
		case <-mh.stop:
			log.Println("stop MsgBlockHandler")
			mh.done <- struct{}{}
			return
		case expectedHeaders = <-mh.expectedBlockHeaders:
			log.Println("Set expected headers")
		case block := <-mh.blocks:
			if err := mh.blockValidator.Validate(block); err != nil {
				log.Printf("block is not valid: %s", err)
				continue
			}
			if err := mh.blockRepository.Save(*block); err != nil {
				log.Println("failed to save block: ", err)
				continue
			}

			currentBlockIndex++
			if currentBlockIndex >= len(expectedHeaders) {
				currentBlockIndex = 0
				expectedHeaders = nil
				mh.notifyProcessedHeaders <- struct{}{}
			}
		}
	}
}

//func nextBlock(headers []p2p.BlockHeader, i int) [32]byte {
//	if i >= len(headers) {
//		return [32]byte{}
//	}
//	return Hash(headers[i])
//}
