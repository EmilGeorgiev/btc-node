package sync

import (
	"log"
	"time"
	
	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type HeadersHandler interface {
	HandleMsgHeaders()
}

type BlockHandler interface {
	HandleBlockMessages()
}

type ChainSync struct {
	headerRequester    HeaderRequester
	headersHandler     HeadersHandler
	blockHandler       BlockHandler
	syncWait           time.Duration
	stop               <-chan struct{}
	processedBlocks    <-chan p2p.MsgBlock
	startFromBlockHash chan<- [32]byte
}

func NewChainSync(hr HeaderRequester, hh HeadersHandler, bh BlockHandler, d time.Duration, s <-chan struct{}) ChainSync {
	return ChainSync{
		headerRequester: hr,
		headersHandler:  hh,
		blockHandler:    bh,
		syncWait:        d,
	}
}

func (cs ChainSync) Start() {
	cs.headersHandler.HandleMsgHeaders()
	cs.blockHandler.HandleBlockMessages()
	go cs.start()
}

func (cs ChainSync) start() {
	timer := time.NewTimer(cs.syncWait)
	for {
		select {
		case <-cs.stop:
			log.Println("stop chain sync iterations")
			return
		case <-cs.processedBlocks:
			timer.Reset(cs.syncWait)
		case <-timer.C:
			timer.Reset(cs.syncWait)
			hash, err := cs.headerRequester.RequestHeadersFromLastBlock()
			if err != nil {
				continue
			}
			cs.startFromBlockHash <- hash
		}
	}
}
