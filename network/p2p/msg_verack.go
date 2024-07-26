package p2p

type MsgVerack struct {
	MessageHeader
}

// NewVerackMsg returns a new 'verack' message.
func NewVerackMsg(network string) (*Message, error) {
	return NewMessage("verack", network, []byte{})
}

type MsgWtxidrelay struct {
	MessageHeader
}

// NewMsgWtxidrelay returns a new 'verack' message.
func NewMsgWtxidrelay(network string) (MsgWtxidrelay, error) {
	msg, _ := NewMessage(CmdWtxidrelay, network, []byte{})
	return MsgWtxidrelay{msg.MessageHeader}, nil
}
