package node

import (
	"encoding/hex"
	"fmt"
	"github.com/EmilGeorgiev/btc-node/network/p2p"
	"math/big"
	"testing"
)

func TestBirsTotarget(t *testing.T) {
	b := BitsToTarget(486604799)
	fmt.Println(b.String())
}

func isOK(hash [32]byte, bits uint32) bool {
	hashBig := new(big.Int).SetBytes(hash[:])
	target := BitsToTarget(bits)

	return hashBig.Cmp(target) <= 0
}

func TestBlockhashIsLessThan(t *testing.T) {
	b, _ := hex.DecodeString("0e0abb91667c0bb906e9ed8bbbfb5876fccb707c2d9e7dab3603b57f41ec431f")
	b1, _ := hex.DecodeString("3a5769fb2126d870aded5fcaced3bc49fa9768436101895931adb5246e41e957")

	rr, _ := hex.DecodeString("c5997d1cad40afec154aa99b8988e97b1f113d8076357a77572455574765a533")

	dd := append(p2p.Reverse([32]byte(b)), p2p.Reverse([32]byte(b1))...)

	actual := fmt.Sprintf("%x", DHash(dd))
	fmt.Println(actual)
	fmt.Printf("%x\n", p2p.Reverse([32]byte(rr)))

}
