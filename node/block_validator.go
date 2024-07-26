package node

import "github.com/EmilGeorgiev/btc-node/network/p2p"

type BlockValidator struct{}

func NewBlockValidator() BlockValidator {
	return BlockValidator{}
}

func (bv BlockValidator) Validate(*p2p.MsgBlock) error {
	return nil
}
