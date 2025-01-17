package sync

import (
	"errors"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"log"
)

var genesisBlockHashHex = "6fe28c0ab6f1b372c1a6a246ae63f74f931e8365e15a089c68d6190000000000"

var GenesisBlockHash = [32]byte{
	0x6f, 0xe2, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72,
	0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
	0x93, 0x1e, 0x83, 0x65, 0xe1, 0x5a, 0x08, 0x9c,
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
	block, err := cs.blockRepository.GetLast()
	var blockHash [32]byte
	if err != nil {
		blockHash = GenesisBlockHash
	} else {
		blockHash = block.GetHash()
	}

	gh, err := p2p.NewMsgGetHeader(cs.network, 1, blockHash, [32]byte{0})
	if err != nil {
		return errors.Join(ErrFailedToCreateMsgGetHeaders, err)
	}

	log.Printf("Find the Last processed block in DB and send it to msg headers handlers: %x\n", p2p.Reverse(blockHash))
	cs.expectedHashes <- blockHash
	cs.outgoingMsgs <- gh
	return nil
}

func (cs HeadersRequester) RequestHeadersFromBlockHash(hash [32]byte) error {
	gh, err := p2p.NewMsgGetHeader(cs.network, 1, hash, [32]byte{0})
	if err != nil {
		return errors.Join(ErrFailedToCreateMsgGetHeaders, err)
	}

	cs.expectedHashes <- hash
	cs.outgoingMsgs <- gh
	return nil
}
