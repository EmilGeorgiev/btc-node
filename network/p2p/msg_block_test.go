package p2p_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/stretchr/testify/require"
	"testing"
)

// 01000000 - version
// 0000000000000000000000000000000000000000000000000000000000000000 - prev block
// 3BA3EDFD7A7B12B27AC72C3E67768F617FC81BC3888A51323A9FB8AA4B1E5E4A - merkle root
// 29AB5F49 - timestamp
// FFFF001D - bits
// 1DAC2B7C - nonce
// 01 - number of transactions
// 01000000 - version
// 01 - input
// 0000000000000000000000000000000000000000000000000000000000000000FFFFFFFF - prev output
// 4D - script length
// 04FFFF001D0104455468652054696D65732030332F4A616E2F32303039204368616E63656C6C6F72206F6E206272696E6B206F66207365636F6E64206261696C6F757420666F722062616E6B73 - scriptsig
// FFFFFFFF - sequence
// 01 - outputs
// 00F2052A01000000 - 50 BTC
// 43 - pk_script length
// 4104678AFDB0FE5548271967F1A67130B7105CD6A828E03909A67962E0EA1F61DEB649F6BC3F4CEF38C4F35504E51EC112DE5C384DF7BA0B8D578A4C702B6BF11D5FAC - pk_script
// 00000000 - lock time
var genesisBlock = "0100000000000000000000000000000000000000000000000000000000000000000000003BA3EDFD7A7B12B27AC72C3E6776" +
	"8F617FC81BC3888A51323A9FB8AA4B1E5E4A29AB5F49FFFF001D1DAC2B7C0101000000010000000000000000000000000000000000000000000" +
	"000000000000000000000FFFFFFFF4D04FFFF001D0104455468652054696D65732030332F4A616E2F32303039204368616E63656C6C6F72206F" +
	"6E206272696E6B206F66207365636F6E64206261696C6F757420666F722062616E6B73FFFFFFFF0100F2052A01000000434104678AFDB0FE554" +
	"8271967F1A67130B7105CD6A828E03909A67962E0EA1F61DEB649F6BC3F4CEF38C4F35504E51EC112DE5C384DF7BA0B8D578A4C702B6BF11D5FAC00000000"

func TestMsgBlock_Unmarshal(t *testing.T) {

	b, _ := hex.DecodeString(genesisBlock)
	fmt.Println(len(genesisBlock))
	block := p2p.MsgBlock{}
	buf := bytes.NewBuffer(b)
	err := binary.NewDecoder(buf).Decode(&block)

	require.NoError(t, err)
	//expect := p2p.MsgBlock{
	//	BlockHeader:  p2p.BlockHeader{},
	//	Transactions: nil,
	//}
	//require.Equal(t, expect, block)
}
