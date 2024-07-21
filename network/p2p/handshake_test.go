package p2p_test

import (
	"fmt"
	"github.com/EmilGeorgiev/btc-node/network/binary"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"github.com/stretchr/testify/require"
	"log"
	"net"
	"testing"
)

func TestCreateHandshake(t *testing.T) {
	validVersionMsg, _ := p2p.NewVersionMsg("mainnet", "test-agent", [4]byte{0x7F, 0x00, 0x00, 0x01}, 3333)
	validVerackMsg, _ := p2p.NewVerackMsg("mainnet")
	pongMsg, _ := p2p.NewPongMsg("mainnet", 11111)
	tests := []struct {
		name      string
		peerAddr  p2p.Addr
		peerNode  dummyNode
		expected  p2p.Handshake
		expectErr bool
	}{
		{
			name:     "successful handshake",
			peerAddr: p2p.Addr{IP: "127.0.0.1", Port: 3333},
			peerNode: dummyNode{
				msgs:     []p2p.Message{*validVersionMsg, *validVerackMsg},
				listenOn: 3333,
			},
			expected: p2p.Handshake{Peer: p2p.Peer{
				Address:   "127.0.0.1:3333",
				Services:  1,
				UserAgent: "test-agent",
				Version:   70015,
			}},
			expectErr: false,
		},
		{
			name:     "remote peer send message version more than once",
			peerAddr: p2p.Addr{IP: "127.0.0.1", Port: 4444},
			peerNode: dummyNode{
				msgs:     []p2p.Message{*validVersionMsg, *validVersionMsg},
				listenOn: 4444,
			},
			expected:  p2p.Handshake{},
			expectErr: true,
		},
		{
			name:     "remote peer send first verack message",
			peerAddr: p2p.Addr{IP: "127.0.0.1", Port: 5555},
			peerNode: dummyNode{
				msgs:     []p2p.Message{*validVersionMsg, *validVersionMsg},
				listenOn: 5555,
			},
			expected:  p2p.Handshake{},
			expectErr: true,
		},
		{
			name:     "remote peer send first unexpected msg pong for handshake",
			peerAddr: p2p.Addr{IP: "127.0.0.1", Port: 5555},
			peerNode: dummyNode{
				msgs:     []p2p.Message{*pongMsg},
				listenOn: 5555,
			},
			expected:  p2p.Handshake{},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(tes *testing.T) {
			// Start a mock server
			err := tt.peerNode.start(tes)
			require.NoError(tes, err)

			actual, err := p2p.CreateHandshake(tt.peerAddr, "mainnet", "test-agent")

			tt.expected.Peer.Connection = actual.Peer.Connection
			require.Equal(tes, tt.expected, actual)
			require.Equal(tes, tt.expectErr, err != nil)
		})
	}
}

type dummyNode struct {
	msgs     []p2p.Message
	listenOn uint16
}

func (dn dummyNode) start(t *testing.T) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", dn.listenOn))
	if err != nil {
		t.Fatalf("failed to create tcp listener. Error: %s", err)
	}

	go func() {
		log.Println("listen for new connections")
		conn, err := listener.Accept()
		if err != nil {
			t.Fatalf("failed to accept connection. Error; %s", err)
		}
		defer listener.Close()
		defer conn.Close()
		// wait the node that opened the connection to initialize handshake workflow. Here we read one byte
		// because this is a dummy node and it only should send list of messages to the connection when receive something
		b := make([]byte, 1)
		conn.Read(b)

		for _, msg := range dn.msgs {
			raw, _ := binary.Marshal(msg)
			if _, err = conn.Write(raw); err != nil {
				t.Fatalf("dummy node failed to write message: %s", err)
			}
		}
	}()

	return nil
}
