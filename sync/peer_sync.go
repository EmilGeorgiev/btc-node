package sync

import (
	"github.com/EmilGeorgiev/btc-node/common"
	"log"
	"math/big"
	"sync/atomic"
	"time"
)

type HeadersOverview struct {
	LastBlockHash [32]byte
	HeadersCount  int64
	CumulativePoW *big.Int
	IsValid       bool
}

type PeerSync struct {
	headerRequester         HeaderRequester
	syncWait                time.Duration
	prevHeadersAreProcessed <-chan struct{}

	prevHeaders <-chan HeadersOverview

	isStarted atomic.Bool
	stop      chan struct{}
	done      chan struct{}
}

func NewPeerSync(hr HeaderRequester, d time.Duration, ph <-chan struct{}, prevHeaders chan HeadersOverview) *PeerSync {
	return &PeerSync{
		headerRequester:         hr,
		syncWait:                d,
		prevHeadersAreProcessed: ph,
		prevHeaders:             prevHeaders,

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

func (cs *PeerSync) GetChainOverview(peerAddr string, ch chan common.ChainOverview) {
	go cs.getChainOverview(peerAddr, ch)
}

var zeroBlockHash = [32]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

func (cs *PeerSync) getChainOverview(peerAddr string, ch chan common.ChainOverview) {
	cho := common.ChainOverview{Peer: peerAddr, CumulativeWork: big.NewInt(0)}
	timer := time.NewTimer(30 * time.Second)
	lastPrevHeaders := HeadersOverview{LastBlockHash: zeroBlockHash}
	_ = cs.headerRequester.RequestHeadersFromBlockHash(zeroBlockHash)

Loop:
	for {
		select {
		case <-cs.stop:
			log.Println("stop chain sync iterations")
			cs.done <- struct{}{}
			return
		case prevHeaders := <-cs.prevHeaders:
			timer.Reset(cs.syncWait)
			if prevHeaders.HeadersCount == 0 {
				break Loop
			}

			if prevHeaders.LastBlockHash == lastPrevHeaders.LastBlockHash {
				continue
			}

			if !prevHeaders.IsValid {
				cho.IsValid = false
				break Loop
			}

			cho.CumulativeWork = cho.CumulativeWork.Add(cho.CumulativeWork, prevHeaders.CumulativePoW)
			cho.NumberOfBlocks += prevHeaders.HeadersCount
			_ = cs.headerRequester.RequestHeadersFromBlockHash(prevHeaders.LastBlockHash)
			lastPrevHeaders = prevHeaders
		case <-timer.C:
			timer.Reset(cs.syncWait)
			_ = cs.headerRequester.RequestHeadersFromBlockHash(lastPrevHeaders.LastBlockHash)
		}
	}

	ch <- cho
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
