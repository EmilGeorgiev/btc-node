package sync

import (
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"log"
)

type BlockValidator interface {
	Validate(p2p.MsgBlock)
}

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
	//var expectedBlocks []p2p.BlockHeader
	//var nextBlockHash [32]byte
	//var index int
	for {
		select {
		case <-mh.stop:
			log.Println("stop MsgBlockHandler")
			return
		case block := <-mh.blocks:
			//if nextBlockHash != block.GetHash() {
			//	log.Printf("unexpected block hash %x. Expect: %x", block.PrevBlockHash, nextBlockHash)
			//	continue
			//}
			mh.blockValidator.Validate(block)
			if err := mh.blockRepository.Save(block); err != nil {
				log.Println("failed to save block: ", err)
				continue
			}
			//index++
			//nextBlockHash = nextBlock(expectedBlocks, index)
			mh.notifyProcessedBlocks <- block
			//case expectedBlocks = <-mh.expectedBlocks:
			//	index = 0
			//	nextBlockHash = nextBlock(expectedBlocks, index)
		}
	}
}

//func nextBlock(headers []p2p.BlockHeader, i int) [32]byte {
//	if i >= len(headers) {
//		return [32]byte{}
//	}
//	return Hash(headers[i])
//}
