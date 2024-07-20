package p2p

// NewVerackMsg returns a new 'verack' message.
func NewVerackMsg(network string) (*Message, error) {
	return NewMessage("verack", network, []byte{})
}
