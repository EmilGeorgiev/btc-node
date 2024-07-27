package node

import (
	"github.com/EmilGeorgiev/btc-node/sync"
	"log"

	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type MsgBlockHandler struct {
	blockRepository       sync.BlockRepository
	blockValidator        sync.BlockValidator
	stop                  chan struct{}
	blocks                <-chan *p2p.MsgBlock
	notifyProcessedBlocks chan<- *p2p.MsgBlock
	done                  chan struct{}
}

func NewMsgBlockHandler(br sync.BlockRepository, bv sync.BlockValidator, blocks <-chan *p2p.MsgBlock, processed chan<- *p2p.MsgBlock) MsgBlockHandler {
	return MsgBlockHandler{
		blockRepository:       br,
		blockValidator:        bv,
		blocks:                blocks,
		notifyProcessedBlocks: processed,

		stop: make(chan struct{}),
		done: make(chan struct{}),
	}
}

func (mh MsgBlockHandler) Start() {
	go mh.handleMsgBlock()
}

func (mh MsgBlockHandler) Stop() {
	close(mh.stop)
	<-mh.done
}

func (mh MsgBlockHandler) handleMsgBlock() {
	for {
		select {
		case <-mh.stop:
			log.Println("stop MsgBlockHandler")
			mh.done <- struct{}{}
			return
		case block := <-mh.blocks:
			if err := mh.blockValidator.Validate(block); err != nil {
				log.Printf("block is not valid: %s", err)
				continue
			}
			if err := mh.blockRepository.Save(*block); err != nil {
				log.Println("failed to save block: ", err)
				continue
			}
			mh.notifyProcessedBlocks <- block
		}
	}
}

//func nextBlock(headers []p2p.BlockHeader, i int) [32]byte {
//	if i >= len(headers) {
//		return [32]byte{}
//	}
//	return Hash(headers[i])
//}
