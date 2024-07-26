package main

//import (
//	"bytes"
//	"crypto/sha256"
//	"encoding/binary"
//	"encoding/hex"
//	"fmt"
//	"log"
//
//	bolt "go.etcd.io/bbolt"
//)
//
//type BlockHeader struct {
//	Version       uint32
//	PrevBlockHash [32]byte
//	MerkleRoot    [32]byte
//	Timestamp     uint32
//	Bits          uint32
//	Nonce         uint32
//}
//
//type Block struct {
//	Header BlockHeader
//	Data   []byte
//}
//
//// Serialize the block header
//func (h *BlockHeader) Serialize() []byte {
//	result := make([]byte, 0, 80)
//	result = append(result, uint32ToBytes(h.Version)...)
//	result = append(result, h.PrevBlockHash[:]...)
//	result = append(result, h.MerkleRoot[:]...)
//	result = append(result, uint32ToBytes(h.Timestamp)...)
//	result = append(result, uint32ToBytes(h.Bits)...)
//	result = append(result, uint32ToBytes(h.Nonce)...)
//	return result
//}
//
//// Convert uint32 to bytes in little-endian
//func uint32ToBytes(val uint32) []byte {
//	result := make([]byte, 4)
//	result[0] = byte(val)
//	result[1] = byte(val >> 8)
//	result[2] = byte(val >> 16)
//	result[3] = byte(val >> 24)
//	return result
//}
//
//// Calculate the block hash
//func (h *BlockHeader) Hash() [32]byte {
//	serializedHeader := h.Serialize()
//	firstHash := sha256.Sum256(serializedHeader)
//	return sha256.Sum256(firstHash[:])
//}
//
//func blockToBytes(block *Block) []byte {
//	headerBytes := block.Header.Serialize()
//	return append(headerBytes, block.Data...)
//}
//
//func main() {
//	// Open the BoltDB database
//	db, err := bolt.Open("blocks.db", 0600, nil)
//	if err != nil {
//		log.Fatal(err)
//	}
//	defer db.Close()
//
//	// Create the "blocks" bucket if it doesn't exist
//	err = db.Update(func(tx *bolt.Tx) error {
//		_, err := tx.CreateBucketIfNotExists([]byte("blocks"))
//		if err != nil {
//			return err
//		}
//		_, err = tx.CreateBucketIfNotExists([]byte("chain"))
//		return err
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Example block header data (version, previous block hash, Merkle root, timestamp, bits, nonce)
//	header := BlockHeader{
//		Version:       1,
//		PrevBlockHash: [32]byte{},
//		MerkleRoot:    [32]byte{0xa4, 0xd6, 0x6f, 0xc5, 0xb1, 0x04, 0x30, 0xfc, 0xfd, 0x14, 0x55, 0x8e, 0x63, 0xd1, 0x9b, 0x64, 0x9a, 0x61, 0xee, 0x95, 0xb7, 0x1b, 0x1b, 0xcc, 0xe9, 0x48, 0xb1, 0xd5, 0x35, 0x83, 0xdb, 0xeb},
//		Timestamp:     486604799,
//		Bits:          0x048eff4f,
//		Nonce:         1,
//	}
//
//	// Example block data
//	blockData, err := hex.DecodeString("010000008e6285267ce431a52e3ef3c46eefc4a144f51195f3bf8489c891ffeb00000000a4d66fc5b10430fcfd14558e63d19b649a61ee95b71b1bcce948b1d53583dbebab176949ffff001d4f7aef04010100000001000000000000000000000000000000000000000000000000000000000000000000ffffffff0704ffff001d0106ffffffff0100f2052a0100000043410419ba01c19d1c68af6bc780289ec1a5d4c181e81089f275325c6e1abc1c2f44c67d99ba9be5d3b9c0b903a8655b853c62717bb99924a1d9bf501d3f9f12b56dc5ac00000000")
//	if err != nil {
//		log.Fatal(err)
//	}
//	block := &Block{
//		Header: header,
//		Data:   blockData,
//	}
//
//	// Calculate the block hash
//	blockHash := block.Header.Hash()
//
//	// Store the block in the "blocks" bucket with the block hash as the key
//	// Also store the previous block hash -> current block hash mapping in the "chain" bucket
//	err = db.Update(func(tx *bolt.Tx) error {
//		bucket := tx.Bucket([]byte("blocks"))
//		err = bucket.Put(blockHash[:], blockToBytes(block))
//		if err != nil {
//			return err
//		}
//		chainBucket := tx.Bucket([]byte("chain"))
//		return chainBucket.Put(block.Header.PrevBlockHash[:], blockHash[:])
//	})
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// Retrieve the latest block hash by traversing the chain from the genesis block
//	var currentHash [32]byte
//	copy(currentHash[:], blockHash[:])
//	var currentBlock *Block
//
//	for {
//		err := db.View(func(tx *bolt.Tx) error {
//			bucket := tx.Bucket([]byte("blocks"))
//			blockData := bucket.Get(currentHash[:])
//			if blockData == nil {
//				return fmt.Errorf("block data not found")
//			}
//
//			currentBlock = &Block{
//				Header: BlockHeader{
//					Version:       binary.LittleEndian.Uint32(blockData[:4]),
//					PrevBlockHash: [32]byte{},
//					MerkleRoot:    [32]byte{},
//					Timestamp:     binary.LittleEndian.Uint32(blockData[68:72]),
//					Bits:          binary.LittleEndian.Uint32(blockData[72:76]),
//					Nonce:         binary.LittleEndian.Uint32(blockData[76:80]),
//				},
//				Data: blockData[80:],
//			}
//			copy(currentBlock.Header.PrevBlockHash[:], blockData[4:36])
//			copy(currentBlock.Header.MerkleRoot[:], blockData[36:68])
//
//			return nil
//		})
//		if err != nil {
//			log.Fatal(err)
//		}
//
//		fmt.Printf("Block Hash: %x\n", currentHash)
//		fmt.Printf("Block Timestamp: %d\n", currentBlock.Header.Timestamp)
//		fmt.Printf("Block Data: %x\n\n", currentBlock.Data)
//
//		if bytes.Equal(currentBlock.Header.PrevBlockHash[:], make([]byte, 32)) {
//			break
//		}
//
//		copy(currentHash[:], currentBlock.Header.PrevBlockHash[:])
//	}
//}
