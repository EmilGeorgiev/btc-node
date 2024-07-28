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

		stop: make(chan struct{}, 10),
		done: make(chan struct{}, 10),
	}
}

func (cs *PeerSync) Start() {
	log.Println("Start PeerSync")
	if !cs.isStarted.Load() {
		log.Println("Start PeerSync 1111111111")
		cs.isStarted.Store(true)
		log.Println("Start PeerSync 22222222222")
		go cs.start()
		log.Println("Start PeerSync 33333333")
		return
	}
	log.Println("Start PeerSync 4444444")
	log.Println("Sync with peer is already started")

}

func (cs *PeerSync) Stop() {
	log.Println("STOP PeerSync")
	if !cs.isStarted.Load() {
		return
	}
	log.Println("STOP PeerSync 111111111")
	cs.stop <- struct{}{}
	log.Println("STOP PeerSync 222222")
	<-cs.done
	log.Println("STOP PeerSync 33333333")
	cs.isStarted.Store(false)
	log.Println("STOP PeerSync 44444444")
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
