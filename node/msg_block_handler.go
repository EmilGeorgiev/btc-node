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
		stop:                 make(chan struct{}, 1000),
		done:                 make(chan struct{}, 1000),
	}
}

func (mh *MsgBlockHandler) Start() {
	log.Println("START MsgBlockhandler")
	if mh.isStarted.Load() {
		log.Println("START MsgBlockhandler 11111111")
		return
	}
	log.Println("START MsgBlockhandler 2222222")
	mh.isStarted.Store(true)
	log.Println("START MsgBlockhandler 33333333")
	go mh.handleMsgBlock()
	log.Println("START MsgBlockhandler 4444444444")
}

func (mh *MsgBlockHandler) Stop() {
	log.Println("Stop MsgBlockhandler")
	if !mh.isStarted.Load() {
		log.Println("Stop MsgBlockhandler 1111111111")
		return
	}
	log.Println("Stop MsgBlockhandler 2222222")
	mh.isStarted.Store(false)
	log.Println("Stop MsgBlockhandler 333333")
	mh.stop <- struct{}{}
	log.Println("Stop MsgBlockhandler 44444444")
	<-mh.done
	log.Println("Stop MsgBlockhandler 55555555555")
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
		case block := <-mh.blocks:
			log.Println("validate block")
			if err := mh.blockValidator.Validate(block); err != nil {
				log.Printf("block is not valid: %s", err)
				continue
			}
			log.Println("save block: ", p2p.Reverse(block.PrevBlockHash[:]))
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
