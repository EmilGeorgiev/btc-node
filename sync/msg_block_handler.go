package sync

import (
	"log"

	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type MsgBlockHandler struct {
	blockRepository       BlockRepository
	blockValidator        BlockValidator
	expectedBlocks        <-chan []p2p.BlockHeader
	stop                  <-chan struct{}
	blocks                <-chan p2p.MsgBlock
	notifyProcessedBlocks chan<- p2p.MsgBlock
}

func NewMsgBlockHandler(br BlockRepository, bv BlockValidator, eb <-chan []p2p.BlockHeader, s <-chan struct{}) MsgBlockHandler {
	return MsgBlockHandler{
		blockRepository: br,
		blockValidator:  bv,
		expectedBlocks:  eb,
		stop:            s,
	}
}

func (mh MsgBlockHandler) HandleMsgBlock() {
	go mh.handleMsgBlock()
}

func (mh MsgBlockHandler) handleMsgBlock() {
	for {
		select {
		case <-mh.stop:
			log.Println("stop MsgBlockHandler")
			return
		case block := <-mh.blocks:
			if err := mh.blockValidator.Validate(block); err != nil {
				log.Printf("block is not valid: %s", err)
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
