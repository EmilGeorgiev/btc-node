package p2p

import "fmt"

const (
	CmdPkgtxns      = "pkgtxns"
	CmdGetpkgtxns   = "getpkgtxns"
	CmdAncpkginfo   = "ancpkginfo"
	CmdSendpackages = "sendpackages"
	CmdPing         = "ping"
	CmdPong         = "pong"

	CmdAddrv2      = "addrv2"
	CmdSendaddrv2  = "sendaddrv2"
	CmdVersion     = "version"
	CmdVerack      = "verack"
	CmdSendcmpct   = "sendcmpct"
	CmdGetheaders  = "getheaders"
	CmdAddr        = "addr"
	CmdInv         = "inv"
	CmdGetdata     = "getdata"
	CmdWtxidrelay  = "wtxidrelay"
	CmdNotfound    = "notfound"
	CmdGetblocks   = "getblocks"
	CmdTx          = "tx"
	CmdBlock       = "block"
	CmdHeaders     = "headers"
	CmdGetadd      = "getadd"
	CmdMempoo      = "mempoo"
	CmdCheckorder  = "checkorder"
	CmdSubmitorder = "submitorder"
	CmdReply       = "reply"
	CmdReject      = "reject"
	CmdFilterload  = "filterload"
	CmdFilteradd   = "filteradd"
	CmdFilterclear = "filterclear"
	CmdMerkleblock = "merkleblock"
	CmdAlert       = "alert"
	CmdSendHeaders = "sendheaders"
	CmdFeefilter   = "feefilter"
	CmdCmpctblock  = "cmpctblock"
	CmdGetblocktxn = "getblocktxn"
	CmdBlocktxn    = "blocktxn"
	CmdMempool     = "mempool"
	CmdGetAddr     = "getaddr"
	CommandLength  = 12
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

var commands = map[string][CommandLength]byte{
	CmdPkgtxns:      newCommand(CmdPkgtxns),
	CmdGetpkgtxns:   newCommand(CmdGetpkgtxns),
	CmdAncpkginfo:   newCommand(CmdAncpkginfo),
	CmdSendpackages: newCommand(CmdSendpackages),
	CmdSendaddrv2:   newCommand(CmdSendaddrv2),
	CmdAddrv2:       newCommand(CmdAddrv2),
	CmdVersion:      newCommand(CmdVersion),
	CmdVerack:       newCommand(CmdVerack),
	CmdPing:         newCommand(CmdPing),
	CmdPong:         newCommand(CmdPong),
	CmdSendcmpct:    newCommand(CmdSendcmpct),
	CmdGetheaders:   newCommand(CmdGetheaders),
	CmdAddr:         newCommand(CmdAddr),
	CmdInv:          newCommand(CmdInv),
	CmdGetdata:      newCommand(CmdGetdata),
	CmdWtxidrelay:   newCommand(CmdWtxidrelay),
	CmdNotfound:     newCommand(CmdNotfound),
	CmdGetblocks:    newCommand(CmdGetblocks),
	CmdTx:           newCommand(CmdTx),
	CmdBlock:        newCommand(CmdBlock),
	CmdHeaders:      newCommand(CmdHeaders),
	CmdGetadd:       newCommand(CmdGetadd),
	CmdMempoo:       newCommand(CmdMempoo),
	CmdCheckorder:   newCommand(CmdCheckorder),
	CmdSubmitorder:  newCommand(CmdSubmitorder),
	CmdReply:        newCommand(CmdReply),
	CmdReject:       newCommand(CmdReject),
	CmdFilterload:   newCommand(CmdFilterload),
	CmdFilteradd:    newCommand(CmdFilteradd),
	CmdFilterclear:  newCommand(CmdFilterclear),
	CmdMerkleblock:  newCommand(CmdMerkleblock),
	CmdAlert:        newCommand(CmdAlert),
	CmdSendHeaders:  newCommand(CmdSendHeaders),
	CmdFeefilter:    newCommand(CmdFeefilter),
	CmdCmpctblock:   newCommand(CmdCmpctblock),
	CmdGetblocktxn:  newCommand(CmdGetblocktxn),
	CmdBlocktxn:     newCommand(CmdBlocktxn),
}

func newCommand(command string) [CommandLength]byte {
	l := len(command)
	if l > CommandLength {
		panic(fmt.Sprintf("command %s is too long\n", command))
	}

	var packed [CommandLength]byte
	buf := make([]byte, CommandLength-l)
	copy(packed[:], append([]byte(command), buf...)[:])

	return packed
}
