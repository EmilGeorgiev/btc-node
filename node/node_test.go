package node

//func TestNode_Start(t *testing.T) {
//	node, err := New("mainnet", "test-agent", 1*time.Minute)
//	require.NoError(t, err)
//
//	networkMock := new(mock)
//	networkMock.Expect().CreateOutgoingHandshake().Returnt(p2p.Handshake{Peer: p})
//	peerSync := new(mock)
//	peerSync.Expect().Start()
//	peerSync.Expect().Start()
//
//
//
//	//startPeer("7777")
//	//startPeer("8888")
//
//	peers := []common.Addr{
//		{IP: "127.0.0.1", Port: 7777},
//		{IP: "127.0.0.1", Port: 8888},
//	}
//	node.ConnectToPeers(peers)
//
//	actual := node.GetServerPeers()
//	expecct := ServerPeer{
//		Sync
//	}
//}
//
//func startPeer(port string) {
//	conn, err := net.Dial("tcp", ":"+port)
//}
