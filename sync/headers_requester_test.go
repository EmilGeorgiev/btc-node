package sync_test

import (
	"fmt"
	"testing"
)

var genesisBlockHash = [32]byte{
	0x6f, 0xe2, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72,
	0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
	0x93, 0x1e, 0x83, 0x65, 0xa1, 0x5a, 0x08, 0x9c,
	0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
}

var genesisBlockHash2 = [32]byte{
	0x6f, 0xe2, 0x8c, 0x0a, 0xb6, 0xf1, 0xb3, 0x72,
	0xc1, 0xa6, 0xa2, 0x46, 0xae, 0x63, 0xf7, 0x4f,
	0x93, 0x1e, 0x83, 0x65, 0xa1, 0x5a, 0x08, 0x9c,
	0x68, 0xd6, 0x19, 0x00, 0x00, 0x00, 0x00, 0x00,
}

func TestTttt(t *testing.T) {
	fmt.Println(genesisBlockHash2)
}

//import (
//	"errors"
//	"github.com/EmilGeorgiev/btc-node/node"
//	"testing"
//
//	"github.com/EmilGeorgiev/btc-node/network/p2p"
//	"github.com/EmilGeorgiev/btc-node/sync"
//	"github.com/golang/mock/gomock"
//	"github.com/stretchr/testify/require"
//)
//
//func TestRequestHeadersFromLastBlock_HappyPath(t *testing.T) {
//	blockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F,
//		0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}
//
//	lastBlock := p2p.MsgBlock{BlockHeader: node.newBlockHeader(blockHash)}
//	msgGetHeaders, _ := p2p.NewMsgGetHeader("mainnet", 1, lastBlock.GetHash(), [32]byte{0})
//
//	ctrl := gomock.NewController(t)
//	blockRepo := sync.NewMockBlockRepository(ctrl)
//	blockRepo.EXPECT().GetLast().Return(lastBlock, nil)
//
//	out := make(chan *p2p.Message, 1)
//	hashes := make(chan [32]byte, 1)
//	hr := sync.NewHeadersRequester("mainnet", blockRepo, out, hashes)
//
//	err := hr.RequestHeadersFromLastBlock()
//	require.NoError(t, err)
//
//	actual := <-out
//	actualHash := <-hashes
//	require.Equal(t, msgGetHeaders, actual)
//	require.Equal(t, lastBlock.GetHash(), actualHash)
//}
//
//func TestRequestHeadersFromLastBlock_WhenGetMsgFromDBFail(t *testing.T) {
//	ctrl := gomock.NewController(t)
//	blockRepo := sync.NewMockBlockRepository(ctrl)
//	blockRepo.EXPECT().GetLast().Return(p2p.MsgBlock{}, errors.New("err"))
//
//	hr := sync.NewHeadersRequester("", blockRepo, nil, nil)
//
//	err := hr.RequestHeadersFromLastBlock()
//	require.NotNil(t, errors.Join(sync.ErrFailedToGetLastBlock, errors.New("err"), err))
//}
