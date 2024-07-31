package sync

import (
	"crypto/sha256"
	"github.com/EmilGeorgiev/btc-node/common"
	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"log"
	"math/big"
	"sync/atomic"
	"time"
)

// RequestedHeaders are the blocks headers that were requested with MsgGetheaders.
type RequestedHeaders struct {
	BlockHeaders  []p2p.BlockHeader
	CumulativePoW *big.Int
	IsValid       bool
}

func (hvr RequestedHeaders) GetLastBlockHeaderHash() [32]byte {
	if len(hvr.BlockHeaders) == 0 {
		return [32]byte{}
	}
	lashHeader := hvr.BlockHeaders[len(hvr.BlockHeaders)-1]
	return Hash(lashHeader)
}

func (hvr RequestedHeaders) GetHeadersNumber() int {
	return len(hvr.BlockHeaders)
}

type PeerSync struct {
	headerRequester  HeaderRequester
	syncWait         time.Duration
	requestedHeaders <-chan RequestedHeaders

	//prevHeaders <-chan RequestedHeaders

	isSyncStarted     atomic.Bool
	isOverviewStarted atomic.Bool
	stop              chan struct{}
	done              chan struct{}
}

func NewPeerSync(hr HeaderRequester, d time.Duration, reqHeaders <-chan RequestedHeaders) *PeerSync {
	return &PeerSync{
		headerRequester:  hr,
		syncWait:         d,
		requestedHeaders: reqHeaders,

		stop: make(chan struct{}, 10),
		done: make(chan struct{}, 10),
	}
}

func (cs *PeerSync) Start() {
	if cs.isSyncStarted.Load() {
		log.Println("PeerSync is already started")
		return
	}

	if cs.isOverviewStarted.Load() {
		log.Println("Peer Chain overview is started. Can't start and the peer sync at the same time")
		return
	}

	cs.isSyncStarted.Store(true)
	log.Println("Start PeerSync")
	go cs.start()
}

func (cs *PeerSync) Stop() {
	if !cs.isSyncStarted.Load() && !cs.isOverviewStarted.Load() {
		log.Println("PeerSync and peer chain overview are not started and is not necessary to be stopped.")
		return
	}
	cs.stop <- struct{}{}
	<-cs.done
	cs.isSyncStarted.Store(false)
	cs.isOverviewStarted.Store(false)
	log.Println("STOP PeerSync")
}

func (cs *PeerSync) StartChainOverview(peerAddr string, ch chan common.ChainOverview) {
	cs.isOverviewStarted.Store(true)
	log.Println("start get chain overview from peersync")
	go cs.getChainOverview(peerAddr, ch)
}

func (cs *PeerSync) getChainOverview(peerAddr string, ch chan common.ChainOverview) {
	defer cs.isOverviewStarted.Store(false)
	cho := common.ChainOverview{Peer: peerAddr, CumulativeWork: big.NewInt(0), IsValid: true}
	timer := time.NewTimer(30 * time.Second)
	var lastHeadersValResult RequestedHeaders
	_ = cs.headerRequester.RequestHeadersFromLastBlock()

Loop:
	for {
		select {
		case <-cs.stop:
			log.Println("stop chain sync iterations")
			cs.done <- struct{}{}
			return
		case headersValResult := <-cs.requestedHeaders:
			timer.Reset(cs.syncWait)

			log.Printf("Last processed block is %x\n", p2p.Reverse(headersValResult.GetLastBlockHeaderHash()))
			log.Println("Common number of processed headers:", cho.NumberOfBlocks)
			//fmt.Printf("Receive processe headers from handler: %#v\n", headersValResult)
			if headersValResult.GetHeadersNumber() == 0 {
				log.Println("Stop the loop in peer sync. get last headers.")
				break Loop
			}

			if headersValResult.GetLastBlockHeaderHash() == lastHeadersValResult.GetLastBlockHeaderHash() {
				log.Println("prev block hash equal to the last processes block hash. Skip this:", p2p.Reverse(lastHeadersValResult.GetLastBlockHeaderHash()))
				continue
			}

			if !headersValResult.IsValid {
				log.Printf("headers are invalid. Lastheader: %x\n", p2p.Reverse(lastHeadersValResult.GetLastBlockHeaderHash()))
				cho.IsValid = false
				break Loop
			}

			cho.CumulativeWork = cho.CumulativeWork.Add(cho.CumulativeWork, headersValResult.CumulativePoW)
			cho.NumberOfBlocks += int64(headersValResult.GetHeadersNumber())
			log.Println("call the RequestHeadersFromBlockHash from PeerSync.getChecinOverview case1:", time.Now().String())
			_ = cs.headerRequester.RequestHeadersFromBlockHash(headersValResult.GetLastBlockHeaderHash())
			lastHeadersValResult = headersValResult
		case <-timer.C:
			if lastHeadersValResult.GetLastBlockHeaderHash() == [32]byte{} {
				_ = cs.headerRequester.RequestHeadersFromLastBlock()
				continue
			}
			log.Printf("Request headers after waiting some seconds: %x\n", p2p.Reverse(lastHeadersValResult.GetLastBlockHeaderHash()))
			timer.Reset(cs.syncWait)
			log.Println("call the RequestHeadersFromBlockHash from PeerSync.getChecinOverview case2:", time.Now().String())
			_ = cs.headerRequester.RequestHeadersFromBlockHash(lastHeadersValResult.GetLastBlockHeaderHash())
		}
	}

	ch <- cho
	log.Println("stop get chain overview")
}

func (cs *PeerSync) start() {
	log.Println("Call RequestHeaders from last block in PeerSync.start:", time.Now().String())
	_ = cs.headerRequester.RequestHeadersFromLastBlock()
	timer := time.NewTimer(30 * time.Second)
	for {
		select {
		case <-cs.stop:
			log.Println("stop chain sync iterations")
			cs.done <- struct{}{}
			return
		case lastSavedHeaders := <-cs.requestedHeaders:
			timer.Reset(cs.syncWait)
			_ = cs.headerRequester.RequestHeadersFromBlockHash(lastSavedHeaders.GetLastBlockHeaderHash())
		case <-timer.C:
			timer.Reset(cs.syncWait)
			cs.requestHeaders()
		}
	}
}

func Hash(bh p2p.BlockHeader) [32]byte {
	b, _ := binary.Marshal(bh)
	firstHash := sha256.Sum256(b[:80])
	return sha256.Sum256(firstHash[:])
}

func (cs *PeerSync) requestHeaders() {
	log.Println("Call RequestHeaders from last block in PeerSync.requestHeaders:", time.Now().String())
	if err := cs.headerRequester.RequestHeadersFromLastBlock(); err != nil {
		log.Printf("failed to Requests headers from peers: %s", err)
		log.Printf("We will tray again after %s", cs.syncWait)
	}
	log.Println("request MSG Headers successfully.")
}
