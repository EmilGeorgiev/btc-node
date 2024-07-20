package p2p

import "fmt"

const (
	cmdPing       = "ping"
	cmdPong       = "pong"
	cmdVersion    = "version"
	commandLength = 12
)

const (
	// Version ...
	Version = 70015

	// SrvNodeNetwork This node can be asked for full blocks instead of just headers.
	SrvNodeNetwork = 1
	// SrvNodeGetUTXO See BIP 0064
	SrvNodeGetUTXO = 2
	// SrvNodeBloom See BIP 0111
	SrvNodeBloom = 4
	// SrvNodeWitness See BIP 0144
	SrvNodeWitness = 8
	// SrvNodeNetworkLimited See BIP 0159
	SrvNodeNetworkLimited = 1024
)

var commands = map[string][commandLength]byte{
	cmdVersion: newCommand(cmdVersion),
}

func newCommand(command string) [commandLength]byte {
	l := len(command)
	if l > commandLength {
		panic(fmt.Sprintf("command %s is too long\n", command))
	}

	var packed [commandLength]byte
	buf := make([]byte, commandLength-l)
	copy(packed[:], append([]byte(command), buf...)[:])

	return packed
}
