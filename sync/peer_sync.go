package sync

import (
	"log"
	"sync/atomic"
	"time"
)

type PeerSync struct {
	headerRequester         HeaderRequester
	syncWait                time.Duration
	prevHeadersAreProcessed <-chan struct{}

	isStarted atomic.Bool
	stop      chan struct{}
	done      chan struct{}
}

func NewPeerSync(hr HeaderRequester, d time.Duration, ph <-chan struct{}) *PeerSync {
	return &PeerSync{
		headerRequester:         hr,
		syncWait:                d,
		prevHeadersAreProcessed: ph,

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
	_ = cs.headerRequester.RequestHeadersFromLastBlock()
	timer := time.NewTimer(30 * time.Second)
	for {
		select {
		case <-cs.stop:
			log.Println("stop chain sync iterations")
			cs.done <- struct{}{}
			return
		case <-cs.prevHeadersAreProcessed:
			timer.Reset(cs.syncWait)
			cs.requestHeaders()
		case <-timer.C:
			timer.Reset(cs.syncWait)
			cs.requestHeaders()
		}
	}
}

func (cs *PeerSync) requestHeaders() {
	log.Println("Request new headers from the last block that the node has.")
	if err := cs.headerRequester.RequestHeadersFromLastBlock(); err != nil {
		log.Printf("failed to Requests headers from peers: %s", err)
		log.Printf("We will tray again after %s", cs.syncWait)
	}
	log.Println("request MSG Headers successfully.")
}
