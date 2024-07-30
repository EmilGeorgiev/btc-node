package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

// Function to compute the SHA256 hash of a byte slice
func sha256Hash(data []byte) []byte {
	hash := sha256.Sum256(data)
	return hash[:]
}

// Function to compute the Merkle Root from a list of transactions
func computeMerkleRoot(transactions []string) string {
	var txHashes [][]byte
	for _, tx := range transactions {
		txBytes, _ := hex.DecodeString(tx)
		txHashes = append(txHashes, sha256Hash(txBytes))
	}

	for len(txHashes) > 1 {
		var newLevel [][]byte
		for i := 0; i < len(txHashes); i += 2 {
			if i+1 == len(txHashes) {
				// Duplicate the last element if the number of elements is odd
				txHashes = append(txHashes, txHashes[i])
			}
			concatenated := append(txHashes[i], txHashes[i+1]...)
			newLevel = append(newLevel, sha256Hash(concatenated))
		}
		txHashes = newLevel
	}

	return hex.EncodeToString(txHashes[0])
}

func main() {
	// Example transactions (these should be in hex format as strings)
	transactions := []string{
		"4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b",
	}

	merkleRoot := computeMerkleRoot(transactions)
	fmt.Println("Computed Merkle Root:", merkleRoot)

	// Compare the computed Merkle Root with the one in the block header
	blockMerkleRoot := "4a5e1e4baab89f3a32518a88c31bc87f618f76673e2cc77ab2127b7afdeda33b"
	if merkleRoot != blockMerkleRoot {
		fmt.Println("Expected: ", blockMerkleRoot)
		fmt.Println("Actual  : ", merkleRoot)
	}
}
