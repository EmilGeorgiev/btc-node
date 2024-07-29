package sync

import (
	"fmt"
	"github.com/EmilGeorgiev/btc-node/common"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
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
	cs.isStarted.Store(true)
	log.Println("start get chain overview from peersync")
	go cs.getChainOverview(peerAddr, ch)
}

var zeroBlockHash = [32]byte{
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
	0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
}

//var genesys = [32]byte{
//	0x00, 0x00, 0x00, 0x00, 0x00, 0x19, 0xd6, 0x68,
//	0x9c, 0x08, 0x5a, 0xe1, 0x65, 0x83, 0x1e, 0x93,
//	0x4f, 0xf7, 0x63, 0xae, 0x46, 0xa2, 0xa6, 0xc1,
//	0x72, 0xb3, 0xf1, 0xb6, 0x0a, 0x8c, 0xe2, 0x6f}

//var GenesisBlockHash = [32]byte{
//	0x6f, 0xe2, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72,
//	0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
//	0x93, 0x1e, 0x83, 0x65, 0xa1, 0x5a, 0x08, 0x9c,
//	0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
//}

func (cs *PeerSync) getChainOverview(peerAddr string, ch chan common.ChainOverview) {
	cho := common.ChainOverview{Peer: peerAddr, CumulativeWork: big.NewInt(0), IsValid: true}
	timer := time.NewTimer(30 * time.Second)
	var lastPrevHeaders HeadersOverview
	fmt.Println("request genesys block.")
	_ = cs.headerRequester.RequestHeadersFromLastBlock()

Loop:
	for {
		select {
		case <-cs.stop:
			log.Println("stop chain sync iterations")
			cs.done <- struct{}{}
			return
		case prevHeaders := <-cs.prevHeaders:
			timer.Reset(cs.syncWait)
			log.Printf("Last processed block is %x\n", p2p.Reverse(prevHeaders.LastBlockHash[:]))
			log.Println("Common number of processed headers:", cho.NumberOfBlocks)
			//fmt.Printf("Receive processe headers from handler: %#v\n", prevHeaders)
			if prevHeaders.HeadersCount == 0 {
				log.Println("Stop the loop in peer sync. get last headers.")
				break Loop
			}

			if prevHeaders.LastBlockHash == lastPrevHeaders.LastBlockHash {
				log.Println("prev block hash equal to the last processes block hash. Skip this:", p2p.Reverse(lastPrevHeaders.LastBlockHash[:]))
				continue
			}

			if !prevHeaders.IsValid {
				log.Println("headers are invalid: ", p2p.Reverse(lastPrevHeaders.LastBlockHash[:]))
				cho.IsValid = false
				break Loop
			}

			cho.CumulativeWork = cho.CumulativeWork.Add(cho.CumulativeWork, prevHeaders.CumulativePoW)
			cho.NumberOfBlocks += prevHeaders.HeadersCount
			_ = cs.headerRequester.RequestHeadersFromBlockHash(prevHeaders.LastBlockHash)
			lastPrevHeaders = prevHeaders
		case <-timer.C:
			log.Printf("Request headers after waiting some seconds: %x\n", p2p.Reverse(lastPrevHeaders.LastBlockHash[:]))
			timer.Reset(cs.syncWait)
			_ = cs.headerRequester.RequestHeadersFromBlockHash(lastPrevHeaders.LastBlockHash)
		}
	}

	ch <- cho
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
