package sync

import (
	"log"
	"time"

	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type ChainSync struct {
	headerRequester    HeaderRequester
	headersHandler     HeadersHandler
	blockHandler       BlockHandler
	syncWait           time.Duration
	processedBlocks    <-chan p2p.MsgBlock
	startFromBlockHash chan<- [32]byte

	stop chan struct{}
	done chan struct{}
}

func NewChainSync(hr HeaderRequester, hh HeadersHandler, bh BlockHandler, d time.Duration, sh chan<- [32]byte, pb <-chan p2p.MsgBlock) ChainSync {
	return ChainSync{
		headerRequester:    hr,
		headersHandler:     hh,
		blockHandler:       bh,
		syncWait:           d,
		startFromBlockHash: sh,
		processedBlocks:    pb,

		stop: make(chan struct{}),
		done: make(chan struct{}, 1),
	}
}

func (cs ChainSync) Start() {
	cs.headersHandler.HandleMsgHeaders()
	cs.blockHandler.HandleBlockMessages()
	go cs.start()
}

func (cs ChainSync) Stop() {
	close(cs.stop)
	<-cs.done
}

func (cs ChainSync) start() {
	timer := time.NewTimer(0 * time.Nanosecond)
	for {
		select {
		case <-cs.stop:
			log.Println("stop chain sync iterations")
			cs.done <- struct{}{}
			return
		case <-cs.processedBlocks:
			timer.Reset(cs.syncWait)
		case <-timer.C:
			log.Println("Start new chain sync iteration.")
			timer.Reset(cs.syncWait)
			hash, err := cs.headerRequester.RequestHeadersFromLastBlock()
			if err != nil {
				continue
			}
			cs.startFromBlockHash <- hash
		}
	}
}