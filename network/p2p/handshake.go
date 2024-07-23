package p2p

import (
	"bytes"
	"fmt"
	"github.com/EmilGeorgiev/btc-node/common"
	"github.com/EmilGeorgiev/btc-node/errors"
	"github.com/EmilGeorgiev/btc-node/network/binary"
	"io"
	"log"
	"net"
)

const minimalSupportedVersion = 70015

func CreateOutgoingHandshake(peerAddr common.Addr, network, userAgent string) (Handshake, error) {
	log.Println("Initialize handshake with peer: ", peerAddr.String())
	conn, err := net.Dial("tcp", peerAddr.String())
	if err != nil {
		msg := fmt.Sprintf("failed to connect to peer: %s", peerAddr.String())
		return Handshake{}, errors.NewE(msg, err, true)
	}

	msg, err := createMsgVersion(peerAddr, network, userAgent)
	if err != nil {
		return Handshake{}, err
	}

	log.Printf("Send MsgVersion to pear: %s: %x\n", peerAddr.String(), msg)
	if _, err = conn.Write(msg); err != nil {
		return Handshake{}, errors.NewE(fmt.Errorf("failed to send MsgVersion to the peer: %s ", peerAddr.String()), err, true)
	}

	msgHeader := make([]byte, MsgHeaderLength)
	versionMsgIsReceived := false
	var handshake Handshake
	for {
		n, err := conn.Read(msgHeader)
		if err != nil {
			return Handshake{}, errors.NewE(
				fmt.Sprintf("receive error while reading from connection from peer: %s.", peerAddr.String()), err, true,
			)
		}

		var header MessageHeader
		if err = binary.NewDecoder(bytes.NewReader(msgHeader[:n])).Decode(&header); err != nil {
			m := fmt.Sprintf("failed to decode header of received messages from peer: %s during the handshake", peerAddr.String())
			return Handshake{}, errors.NewE(m, err, true)
		}

		if err = header.Validate(); err != nil {
			m := fmt.Sprintf("Error while validate message header from peer: %s", peerAddr.String())
			return Handshake{}, errors.NewE(m, err, true)
		}

		switch header.CommandString() {
		case "version":
			log.Println("receive msg version from peer: ", peerAddr.String())
			if versionMsgIsReceived {
				m := fmt.Sprintf("message version is received twice during the handhsake with peer: %s which violate protocol", peerAddr.String())
				return Handshake{}, errors.NewE(m)
			}
			versionMsgIsReceived = true
			handshake, err = handleVersion(header, conn)
			if err != nil {
				return Handshake{}, err
			}
		case "verack":
			log.Println("receive msg verack")
			if versionMsgIsReceived {
				return handshake, nil
			}
			log.Println("verack is received before msg version")
			m := fmt.Errorf("unexpected message of type verack is received before msg  version during the handshake with peer: %s", peerAddr.String())
			return Handshake{}, errors.NewE(m)
		default:
			log.Println("receive unexpected message:", header.CommandString())
			m := fmt.Errorf("unexpected message of type %s is received during the handshake with peer: %s", header.CommandString(), peerAddr.String())
			return Handshake{}, errors.NewE(m)
		}
	}
}

func handleVersion(msgHeader MessageHeader, conn net.Conn) (Handshake, error) {
	var version MsgVersion

	lr := io.LimitReader(conn, int64(msgHeader.Length))
	if err := binary.NewDecoder(lr).Decode(&version); err != nil {
		return Handshake{}, errors.NewE(
			fmt.Sprintf("failed decode MsgVersion from peer: %s.", conn.RemoteAddr().String()), err, true)
	}

	peer := Peer{
		Address:    conn.RemoteAddr().String(),
		Connection: conn,
		Services:   version.Services,
		UserAgent:  version.UserAgent.String,
		Version:    version.Version,
	}

	if minimalSupportedVersion > version.Version {
		return Handshake{}, errors.NewE(
			fmt.Sprintf("peer: %s support to old protocol version: %d. Minimum supported version is: %d",
				peer.Address, version.Version, minimalSupportedVersion))
	}

	verack, err := NewVerackMsg("mainnet")
	if err != nil {
		return Handshake{}, err
	}

	msg, err := binary.Marshal(verack)
	if err != nil {
		return Handshake{}, errors.NewE(fmt.Sprintf("failed to marshal verack msg for peer %s", peer.Address), err)
	}

	fmt.Printf("Send verack message to peer")
	if _, err := conn.Write(msg); err != nil {
		return Handshake{}, errors.NewE(
			fmt.Sprintf("failed to send verack message through conn to peer: %s", peer.Address), err, true)
	}

	return Handshake{Peer: peer}, nil
}

type Handshake struct {
	Peer Peer
}

// Addr ...
//type Addr struct {
//	IP   string
//	Port int64
//}

//func (a Addr) String() string {
//	return fmt.Sprintf("%s:%d", a.IP, a.Port)
//}

func createMsgVersion(peerAddr common.Addr, network, userAgent string) ([]byte, error) {
	ip := net.ParseIP(peerAddr.IP)
	a := IPv4{}
	copy(a[:], ip.To4())
	version, err := NewVersionMsg(network, userAgent, a, uint16(peerAddr.Port))
	if err != nil {
		return nil, err
	}

	b, err := binary.Marshal(version)
	if err != nil {
		return nil, errors.NewE("failed to marshal MsgVersion.", err)
	}
	return b, nil
}

// message headrs that contains list woth block headers
// header:  f9beb4d9686561646572730000000000d3780200a65b2b77
// []blockheaders: fdd007010000006fe28c0ab6f1b372c1a6a246ae63f74f931e8365e15a089c68d6190000000000982051fd1e4ba744bbbe680e1fee14677ba1a3c3540bf7b1cdb606e857233e0e61bc6649ffff001d01e3629900010000004860eb18bf1b1620e37e9490fc8a427514416fd75159ab86688e9a8300000000d5fdcc541e25de1c7a5addedf24858b8bb665c9f36ef744ee42c316022c90f9bb0bc6649ffff001d08d2bd610001000000bddd99ccfda39da1b108ce1a5d70038d0a967bacb68b6b63065f626a0000000044f672226090d85db9a9f2fbfe5f0f9609b387af7be5b7fbb7a1767c831c9e995dbe6649ffff001d05e0ed6d00010000004944469562ae1c2c74d9a535e00b6f3e40ffbad4f2fda3895501b582000000007a06ea98cd40ba2e3288262b28638cec5337c1456aaf5eedc8e9e5a20f062bdf8cc16649ffff001d2bfee0a9000100000085144a84488ea88d221c8bd6c059da090e88f8a2c99690ee55dbba4e00000000e11c48fecdd9e72510ca84f023370c9a38bf91ac5cae88019bee94d24528526344c36649ffff001d1d03e4770001000000fc33f596f822a0a1951ffdbf2a897b095636ad871707bf5d3162729b00000000379dfb96a5ea8c81700ea4ac6b97ae9a9312b2d4301a29580e924ee6761a2520adc46649ffff001d189c4c9700010000008d778fdc15a2d3fb76b7122a3b5582bea4f21f5a0c693537e7a03130000000003f674005103b42f984169c7d008370967e91920a6a5d64fd51282f75bc73a68af1c66649ffff001d39a59c8600010000004494c8cf4154bdcc0720cd4a59d9c9b285e4b146d45f061d2b6c967100000000e3855ed886605b6d4a99d5fa2ef2e9b0b164e63df3c4136bebf2d0dac0f1f7a667c86649ffff001d1c4b56660001000000c60ddef1b7618ca2348a46e868afc26e3efc68226c78aa47f8488c4000000000c997a5e56e104102fa209c6a852dd90660a20b2d9c352423edce25857fcd37047fca6649ffff001d28404f5300010000000508085c47cc849eb80ea905cc7800a3be674ffc57263cf210c59d8d00000000112ba175a1e04b14ba9e7ea5f76ab640affeef5ec98173ac9799a852fa39add320cd6649ffff001d1e2de5650001000000e915d9a478e3adf3186c07c61a22228b10fd87df343c92782ecc052c000000006e06373c80de397406dc3d19c90d71d230058d28293614ea58d6a57f8f5d32f8b8ce6649ffff001d173807f800010000007330d7adf261c69891e6ab08367d957e74d4044bc5d9cd06d656be9700000000b8c8754fabb0ffeb04ca263a1368c39c059ca0d4af3151b876f27e197ebb963bc8d06649ffff001d3f596a0c00010000005e2b8043bd9f8db558c284e00ea24f78879736f4acd110258e48c2270000000071b22998921efddf90c75ac3151cacee8f8084d3e9cb64332427ec04c7d562994cd16649ffff001d37d1ae86000100000089304d4ba5542a22fb616d1ca019e94222ee45c1ad95a83120de515c00000000560164b8bad7675061aa0f43ced718884bdd8528cae07f24c58bb69592d8afe185d36649ffff001d29cbad240001000000378a6f6593e2f0251132d96616e837eb6999bca963f6675a0c7af180000000000d080260d107d269ccba9247cfc64c952f1d13514b49e9f1230b3a197a8b7450fa276849ffff001d38d8fb9800010000007384231257343f2fa3c55ee69ea9e676a709a06dcfd2f73e8c2c32b300000000442ee91b2b999fb15d61f6a88ecf2988e9c8ed48f002476128e670d3dac19fe706286849ffff001d049e12d60001000000f5c46c41c30df6aaff3ae9f74da83e4b1cffdec89c009b39bb254a17000000005d6291c35a88fd9a3aef5843124400936fbf2c9166314addcaf5678e55b7e0a30f2c6849ffff001d07608493000100000009f8fd6ba6f0b6d5c207e8fcbcf50f46876a5deffbac4701d7d0f13f0000000023ca63b851cadfd7099ae68eb22147d09394adb72a78e86b69c42deb6df225f92e2e6849ffff001d323741f20001000000161126f0d39ec082e51bbd29a1dfb40b416b445ac8e493f88ce993860000000030e2a3e32abf1663a854efbef1b233c67c8cdcef5656fe3b4f28e52112469e9bae306849ffff001d16d1b42d00010000006f187fddd5e28aa1b4065daa5d9eae0c487094fb20cf97ca02b81c84000000005b7b25b51797f83192f9fd2c3871bfb27570a7d6b56d3a50760613d1a2fc1aeeab346849ffff001d36d950710001000000d7c834e8ea05e2c2fddf4d82faf4c3e921027fa190f1b8372a7aa96700000000b41092b870cc096070ff3212c207c0881e3a2abafc1b92507941b4ef705917e0d9366849ffff001d2bd021d600010000004f29f31e6dac13710ae72d54278b5c97ff6c1646e95b27d14263016f000000004349d6a4e94f05a736ac830754e76dfdf7f140c331f316d1a278517e1daf2e9e6b3a6849ffff001d28140f6200010000003b5e5b888c8c3da0f1d6c3969e63a7a9c1215a3360c8107a428db598000000008c4cc1b42c9dab1973890ecdfdee032079ed39892ad53a6546844d237634cfe1fb3a6849ffff001d255ab455000100000082219cebbdc9bcb715efee535c13a44447e99dfaff6d552e9839d30c000000003e75f63c634ed5fb3d8e21de5fe143cfa63c8018fce0fa26cbc628378b9bc343953d6849ffff001d27ba00b100010000005f411e0d7783fc274b4fea8597209d31d4a511e887a489cebb1f05fc00000000be2123ad48038313b8b726a51cb080bb5a8b81c4166401493b017d2d33520f9b063f6849ffff001d2337f13100010000002620766fa24558ad47e3a9623cd17ff4623668768dbea19ed5a1358e00000000dc1490b5ba227b1adbb2513f74e0252e8fe68b6c7de74c1a22adb63b14e8c16712466849ffff001d344eb75c00010000009810f0fa1817a4d2d371a069addaafab2ca99887abcc5bd2528e434100000000654f005a6e4b4b57b42343fb0e47f32079b4ebfe643c2ea4ea20e46c3af00c238d466849ffff001d364c8cb3000100000081203520416c370fde3d6d46e82ed4332b5035bfba848ff97207357100000000bdaed84e0cbab735880d4763a1eb2df1ecd59dc261f3446db37bed5b6ccb99f331bf6849ffff001d2e5bd48e00010000004409709aff1b155be4f7a9ccef6121345050be74b4bad1d330940dbb00000000ec77d34cb2f84f3447c37ec1b4476e044e88478378998bd55d031f58f4e261c35fbf6849ffff001d32cb39a00001000000cb9ba5a45252b335fe47a099c8935d01ff8eef2e598c2051631b7ac50000000031534f7571b5ea98c1318eed04937d6ff16582ba72c53552581c40828b6ce2f5cac16849ffff001d080315e80001000000db643f0756bb4f6b25ce4a475b533d9ef75cd536e72df664fb9c91bc00000000cb527bd29495c02c9d6515de91ef264df333447e48ef730f3b66ffa8db3eb38630c46849ffff001d155dbb2a0001000000c4d369b723c2cf9be33cf00deb1dbfea0c8ccd12c415f29434ff009700000000c9c0fd0ae7b7973c42fc9e3dddc967b6e309570b720ff15414c08365f005992be3c56849ffff001d08e1c00d0001000000e3f6664d5af37062b934f983ed1033e2011b42c9b04735276c7ccbe5000000001012aaab3e3bffd34055aaa157bf78792d5c18f085635eda7046d89c08a0eabde3c86849ffff001d228c22400001000000627985c0fc1a71e052a5af9420c9b99845432ae099f27a3dea7370a80000000074549b3151d6dd4ce77419d01710921b3211ed3280bf2e3af2c1f1a820063b2272ca6849ffff001d2243c02400010000008f31b4c405cfc212fa4e62840dc8d0
