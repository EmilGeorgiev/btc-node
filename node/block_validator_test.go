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
	b, _ := hex.DecodeString("0000000000000000000aef8215165b0d10dbb7ce3374871bcd3081ddd2392882")
	fmt.Println(isOK([32]byte(p2p.Reverse(b)), 386673224))

}
