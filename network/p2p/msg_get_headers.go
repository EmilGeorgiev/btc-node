package p2p

type MsgGetHeader struct {
	Version    uint32
	HashCount  uint8
	StartBlock [32]byte
	StopBlock  [32]byte
}

func NewMsgGetHeader(network string, hashCount uint8, startBlock, stopBlock [32]byte) (*Message, error) {
	payload := MsgGetHeader{
		Version:    Version,
		HashCount:  hashCount,
		StartBlock: startBlock,
		StopBlock:  stopBlock,
	}

	return NewMessage(cmdGetheaders, network, payload)
}
