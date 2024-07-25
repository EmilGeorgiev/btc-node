package sync

import (
	"fmt"

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
		return [32]byte{}, fmt.Errorf("failed to get latest block, %s", err)
	}

	gh, err := p2p.NewMsgGetHeader(cs.network, 1, block.GetHash(), [32]byte{0})
	if err != nil {
		return [32]byte{}, fmt.Errorf("failed to create msg GetHeaders: %s", err)
	}

	if err = cs.msgSender.SendMsg(*gh); err != nil {
		return [32]byte{}, fmt.Errorf("failed to send msg GetHeaders: %s", err)
	}

	return block.GetHash(), nil
}
