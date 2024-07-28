package sync

import (
	"errors"
	"fmt"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

//	var genesisBlockHash = [32]byte{
//		0x00, 0x00, 0x00, 0x00, 0x00, 0x19, 0xd6, 0x68,
//		0x9c, 0x08, 0x5a, 0xe1, 0x65, 0x83, 0x1e, 0x93,
//		0x4f, 0xf7, 0x63, 0xae, 0x46, 0xa2, 0xa6, 0xc1,
//		0x72, 0xb3, 0xf1, 0xb6, 0x0a, 0x8c, 0xe2, 0x6f,
//	}
var genesisBlockHash = [32]byte{
	0x6f, 0xe2, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72,
	0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
	0x93, 0x1e, 0x83, 0x65, 0xa1, 0x5a, 0x08, 0x9c,
	0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
}

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
	fmt.Println("In request header")
	block, err := cs.blockRepository.GetLast()
	var blockHash [32]byte
	if err != nil {
		blockHash = genesisBlockHash
	} else {
		blockHash = block.GetHash()
	}

	gh, err := p2p.NewMsgGetHeader(cs.network, 1, blockHash, [32]byte{0})
	if err != nil {
		return errors.Join(ErrFailedToCreateMsgGetHeaders, err)
	}

	cs.expectedHashes <- blockHash
	cs.outgoingMsgs <- gh
	return nil
}
