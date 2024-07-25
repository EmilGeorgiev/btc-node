package sync

import (
	"errors"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
)

type HeadersRequester struct {
	network         string
	blockRepository BlockRepository
	msgSender       MsgSender
}

func NewHeadersRequester(n string, br BlockRepository, ms MsgSender) HeadersRequester {
	return HeadersRequester{
		network:         n,
		blockRepository: br,
		msgSender:       ms,
	}
}

func (cs HeadersRequester) RequestHeadersFromLastBlock() ([32]byte, error) {
	block, err := cs.blockRepository.GetLast()
	if err != nil {
		return [32]byte{}, errors.Join(ErrFailedToGetLastBlock, err)
	}

	gh, err := p2p.NewMsgGetHeader(cs.network, 1, block.GetHash(), [32]byte{0})
	if err != nil {
		return [32]byte{}, errors.Join(ErrFailedToCreateMsgGetHeaders, err)
	}

	if err = cs.msgSender.SendMsg(*gh); err != nil {
		return [32]byte{}, errors.Join(ErrFailedToSendMsgGetHeaders, err)
	}

	return block.GetHash(), nil
}
