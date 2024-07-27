package sync

import (
	"log"
	"sync/atomic"
	"time"

	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type PeerSync struct {
	headerRequester HeaderRequester
	syncWait        time.Duration
	processedBlocks <-chan *p2p.MsgBlock

	isStarted atomic.Bool
	stop      chan struct{}
	done      chan struct{}
}

func NewPeerSync(hr HeaderRequester, d time.Duration, pb <-chan *p2p.MsgBlock) *PeerSync {
	return &PeerSync{
		headerRequester: hr,
		syncWait:        d,
		processedBlocks: pb,

		stop: make(chan struct{}),
		done: make(chan struct{}, 1),
	}
}

func (cs *PeerSync) Start() {
	if !cs.isStarted.Load() {
		cs.isStarted.Store(true)
		go cs.start()
		return
	}
	log.Println("Sync with peer is already started")

}

func (cs *PeerSync) Stop() {
	if !cs.isStarted.Load() {
		return
	}
	cs.stop <- struct{}{}
	<-cs.done
	cs.isStarted.Store(false)
}

func (cs *PeerSync) start() {
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
			if err := cs.headerRequester.RequestHeadersFromLastBlock(); err != nil {
				log.Printf("failed to Requests headers from peers: %s", err)
				log.Printf("We will tray again after %s", cs.syncWait)
				continue
			}
			log.Println("request MSGHeader succsessfully")
		}
	}
}
