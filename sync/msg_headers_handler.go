package sync

import (
	"crypto/sha256"

	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type HeaderRequester interface {
	RequestHeadersFromLastBlock() ([32]byte, error)
}

type BlockRepository interface {
	Save(block p2p.MsgBlock) error
	Get(key [32]byte) (p2p.MsgBlock, error)
	GetLast() (p2p.MsgBlock, error)
}

type MsgSender interface {
	SendMsg(message p2p.Message) error
}

type MsgHeadersHandler struct {
	network                       string
	headerRequester               HeaderRequester
	blockRepository               BlockRepository
	msgSender                     MsgSender
	headers                       <-chan p2p.MsgHeaders
	blocks                        <-chan p2p.MsgBlock
	blockHashes                   <-chan [32]byte
	notifyForExpectedBlocks       chan<- []p2p.BlockHeader
	stop                          chan struct{}
	signalStartOfNewSyncIteration chan struct{}
}

func NewMsgHeaderHandler(n string, br BlockRepository, ms MsgSender, h <-chan p2p.MsgHeaders, b <-chan p2p.MsgBlock) ChainSyncInitializer {
	return ChainSyncInitializer{
		network:         n,
		blockRepository: br,
		msgSender:       ms,
		headers:         h,
		blocks:          b,
		expectedHeaders: make(chan p2p.MsgGetHeader, 1),
	}
}

func (mh MsgHeadersHandler) HandleMsgHeaders() {
	go mh.handleHeaders()
}

func (mh MsgHeadersHandler) handleHeaders() {
	var expectBlockHeadersStartedFromHash [32]byte
	for {
		select {
		case <-mh.stop:
			return
		case expectBlockHeadersStartedFromHash = <-mh.blockHashes:
		case msgH := <-mh.headers: // handle MsgHeaders
			if msgH.BlockHeaders[0].PrevBlockHash != expectBlockHeadersStartedFromHash {
				continue
			}

			inv := make([]p2p.InvVector, len(msgH.BlockHeaders))
			for i := 0; i < len(msgH.BlockHeaders); i++ {
				inv[i] = p2p.InvVector{Type: 2, Hash: Hash(msgH.BlockHeaders[i])}
			}

			msgGetdata := p2p.MsgGetData{Count: p2p.VarInt(len(msgH.BlockHeaders)), Inventory: inv}
			msg, _ := p2p.NewMessage(p2p.CmdGetdata, mh.network, msgGetdata)
			if err := mh.msgSender.SendMsg(*msg); err != nil {

			}

			mh.notifyForExpectedBlocks <- msgH.BlockHeaders
		}
	}
}

func (cs ChainSyncInitializer) Start() {
	for {
		select {
		case <-cs.stop:
			return
		default:
			expectedHeader := cs.RequestHeadersFromLastBlock()
			HandleMshHeaders(expectedHeader.PrevBlockhash)
			handleBlockMsgs()
		}

	}
}

func Hash(bh p2p.BlockHeader) [32]byte {
	b, _ := binary.Marshal(bh)
	firstHash := sha256.Sum256(b)
	return sha256.Sum256(firstHash[:])
}
