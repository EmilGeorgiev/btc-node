package p2p

import "fmt"

const (
	cmdPkgtxns      = "pkgtxns"
	cmdGetpkgtxns   = "getpkgtxns"
	cmdAncpkginfo   = "ancpkginfo"
	cmdSendpackages = "sendpackages"
	cmdPing         = "ping"
	cmdPong         = "pong"

	CmdAddrv2      = "addrv2"
	cmdSendaddrv2  = "sendaddrv2"
	cmdVersion     = "version"
	cmdVerack      = "verack"
	cmdSendcmpct   = "sendcmpct"
	cmdGetheaders  = "getheaders"
	cmdAddr        = "addr"
	cmdInv         = "inv"
	CmdGetdata     = "getdata"
	CmdWtxidrelay  = "wtxidrelay"
	cmdNotfound    = "notfound"
	cmdGetblocks   = "getblocks"
	cmdTx          = "tx"
	cmdBlock       = "block"
	cmdHeaders     = "headers"
	cmdGetadd      = "getadd"
	cmdMempoo      = "mempoo"
	cmdCheckorder  = "checkorder"
	cmdSubmitorder = "submitorder"
	cmdReply       = "reply"
	cmdReject      = "reject"
	cmdFilterload  = "filterload"
	cmdFilteradd   = "filteradd"
	cmdFilterclear = "filterclear"
	cmdMerkleblock = "merkleblock"
	cmdAlert       = "alert"
	cmdSendHeaders = "sendheaders"
	cmdFeefilter   = "feefilter"
	cmdCmpctblock  = "cmpctblock"
	cmdGetblocktxn = "getblocktxn"
	cmdBlocktxn    = "blocktxn"
	cmdMempool     = "mempool"
	cmdGetAddr     = "getaddr"
	commandLength  = 12
)

const (
	// Version ...
	Version = 70016

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
	cmdPkgtxns:      newCommand(cmdPkgtxns),
	cmdGetpkgtxns:   newCommand(cmdGetpkgtxns),
	cmdAncpkginfo:   newCommand(cmdAncpkginfo),
	cmdSendpackages: newCommand(cmdSendpackages),
	cmdSendaddrv2:   newCommand(cmdSendaddrv2),
	CmdAddrv2:       newCommand(CmdAddrv2),
	cmdVersion:      newCommand(cmdVersion),
	cmdVerack:       newCommand(cmdVerack),
	cmdPing:         newCommand(cmdPing),
	cmdPong:         newCommand(cmdPong),
	cmdSendcmpct:    newCommand(cmdSendcmpct),
	cmdGetheaders:   newCommand(cmdGetheaders),
	cmdAddr:         newCommand(cmdAddr),
	cmdInv:          newCommand(cmdInv),
	CmdGetdata:      newCommand(CmdGetdata),
	CmdWtxidrelay:   newCommand(CmdWtxidrelay),
	cmdNotfound:     newCommand(cmdNotfound),
	cmdGetblocks:    newCommand(cmdGetblocks),
	cmdTx:           newCommand(cmdTx),
	cmdBlock:        newCommand(cmdBlock),
	cmdHeaders:      newCommand(cmdHeaders),
	cmdGetadd:       newCommand(cmdGetadd),
	cmdMempoo:       newCommand(cmdMempoo),
	cmdCheckorder:   newCommand(cmdCheckorder),
	cmdSubmitorder:  newCommand(cmdSubmitorder),
	cmdReply:        newCommand(cmdReply),
	cmdReject:       newCommand(cmdReject),
	cmdFilterload:   newCommand(cmdFilterload),
	cmdFilteradd:    newCommand(cmdFilteradd),
	cmdFilterclear:  newCommand(cmdFilterclear),
	cmdMerkleblock:  newCommand(cmdMerkleblock),
	cmdAlert:        newCommand(cmdAlert),
	cmdSendHeaders:  newCommand(cmdSendHeaders),
	cmdFeefilter:    newCommand(cmdFeefilter),
	cmdCmpctblock:   newCommand(cmdCmpctblock),
	cmdGetblocktxn:  newCommand(cmdGetblocktxn),
	cmdBlocktxn:     newCommand(cmdBlocktxn),
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
