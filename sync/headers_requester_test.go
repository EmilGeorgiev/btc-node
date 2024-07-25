package sync_test

import (
	"errors"
	"testing"

	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/EmilGeorgiev/btc-node/sync"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestRequestHeadersFromLastBlock_HappyPath(t *testing.T) {
	blockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F,
		0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	lastBlock := p2p.MsgBlock{BlockHeader: newBlockHeader(blockHash)}
	msgGetHeaders, _ := p2p.NewMsgGetHeader("mainnet", 1, lastBlock.GetHash(), [32]byte{0})

	ctrl := gomock.NewController(t)
	blockRepo := sync.NewMockBlockRepository(ctrl)
	blockRepo.EXPECT().GetLast().Return(lastBlock, nil)
	msgSender := sync.NewMockMsgSender(ctrl)
	msgSender.EXPECT().SendMsg(*msgGetHeaders, "8.8.8.8/32")

	hr := sync.NewHeadersRequester("mainnet", blockRepo, msgSender)

	actual, err := hr.RequestHeadersFromLastBlock("8.8.8.8/32")

	require.NoError(t, err)
	require.Equal(t, lastBlock.GetHash(), actual)
}

func TestRequestHeadersFromLastBlock_WhenGetMsgFromDBFail(t *testing.T) {
	ctrl := gomock.NewController(t)
	blockRepo := sync.NewMockBlockRepository(ctrl)
	blockRepo.EXPECT().GetLast().Return(p2p.MsgBlock{}, errors.New("err"))

	hr := sync.NewHeadersRequester("", blockRepo, nil)

	_, err := hr.RequestHeadersFromLastBlock("127.0.0.1/32")
	require.NotNil(t, errors.Join(sync.ErrFailedToGetLastBlock, errors.New("err"), err))
}

func TestRequestHeadersFromLastBlock_WhenSendMsgFail(t *testing.T) {
	blockHash := [32]byte{0x3B, 0xA3, 0xED, 0xFD, 0x7A, 0x7B, 0x12, 0xB2, 0x7A, 0xC7, 0x2C, 0x3E, 0x67, 0x76, 0x8F,
		0x61, 0x7F, 0xC8, 0x1B, 0xC3, 0x88, 0x8A, 0x51, 0x32, 0x3A, 0x9F, 0xB8, 0xAA, 0x4B, 0x1E, 0x5E, 0x4A}

	lastBlock := p2p.MsgBlock{BlockHeader: newBlockHeader(blockHash)}
	msgGetHeaders, _ := p2p.NewMsgGetHeader("mainnet", 1, lastBlock.GetHash(), [32]byte{0})

	ctrl := gomock.NewController(t)
	blockRepo := sync.NewMockBlockRepository(ctrl)
	blockRepo.EXPECT().GetLast().Return(lastBlock, nil)
	msgSender := sync.NewMockMsgSender(ctrl)
	msgSender.EXPECT().SendMsg(*msgGetHeaders, "127.0.0.1/32").Return(errors.New("err"))

	hr := sync.NewHeadersRequester("mainnet", blockRepo, msgSender)

	_, err := hr.RequestHeadersFromLastBlock("127.0.0.1/32")
	require.Equal(t, errors.Join(sync.ErrFailedToSendMsgGetHeaders, errors.New("err")), err)
}
