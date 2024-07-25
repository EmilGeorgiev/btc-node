package sync

import (
	"log"

	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type MsgBlockHandler struct {
	blockRepository       BlockRepository
	blockValidator        BlockValidator
	stop                  chan struct{}
	blocks                <-chan p2p.MsgBlock
	notifyProcessedBlocks chan<- p2p.MsgBlock
	done                  chan struct{}
}

func NewMsgBlockHandler(br BlockRepository, bv BlockValidator, blocks <-chan p2p.MsgBlock, s chan struct{}, notify chan<- p2p.MsgBlock) MsgBlockHandler {
	return MsgBlockHandler{
		blockRepository:       br,
		blockValidator:        bv,
		blocks:                blocks,
		stop:                  s,
		notifyProcessedBlocks: notify,
		done:                  make(chan struct{}),
	}
}

func (mh MsgBlockHandler) HandleMsgBlock() {
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
			close(mh.done)
			return
		case block := <-mh.blocks:
			if err := mh.blockValidator.Validate(block); err != nil {
				log.Printf("block is not valid: %s", err)
				continue
			}
			if err := mh.blockRepository.Save(block); err != nil {
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
