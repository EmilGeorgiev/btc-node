package sync

import (
	"errors"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type HeadersRequester struct {
	network         string
	blockRepository BlockRepository

	// used to queue messages that needs to be send to the peer
	outgoingMsgs chan<- *p2p.Message

	// used to notify MsgHeaders handler what to expect in Header msg. It should reject
	// all headers messages in which the first block header's previous block hash is not
	// the one that is send through this channel
	expectedHashes chan<- [32]byte
}

func NewHeadersRequester(n string, br BlockRepository, out chan<- *p2p.Message, h chan<- [32]byte) HeadersRequester {
	return HeadersRequester{
		network:         n,
		blockRepository: br,
		outgoingMsgs:    out,
		expectedHashes:  h,
	}
}

func (cs HeadersRequester) RequestHeadersFromLastBlock() error {
	block, err := cs.blockRepository.GetLast()
	if err != nil {
		return errors.Join(ErrFailedToGetLastBlock, err)
	}

	gh, err := p2p.NewMsgGetHeader(cs.network, 1, block.GetHash(), [32]byte{0})
	if err != nil {
		return errors.Join(ErrFailedToCreateMsgGetHeaders, err)
	}

	cs.expectedHashes <- block.GetHash()
	cs.outgoingMsgs <- gh
	return nil
}
