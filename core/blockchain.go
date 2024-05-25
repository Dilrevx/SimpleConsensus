package core

import (
	"math/rand"
	"nis3607/mylogger"
	"sync"
	"time"
)

// Contains seq and Data(dynamic length, not hard-coded)
type Block struct {
	Seq  uint64
	Data []byte
}

// Local representation of a blockchain
// Contains []*Block, blocksMap, keysMap
// Thread-safe with a mutex
//
// Blockchain have a independent logger
type BlockChain struct {
	Id        uint8    // node id
	BlockSize uint64   // block size
	Blocks    []*Block // block array
	BlocksMap map[string]*Block
	KeysMap   map[*Block]string
	logger    *mylogger.MyLogger
	mu        sync.Mutex
}

// Initialize chain with 1024 blocks, empty blocksMap and keysMap
func InitBlockChain(id uint8, blocksize uint64) *BlockChain {
	blocks := make([]*Block, 1024)
	blocksMap := make(map[string]*Block)
	keysMap := make(map[*Block]string)
	//Generate gensis block
	blockChain := &BlockChain{
		Id:        id,
		BlockSize: blocksize,
		Blocks:    blocks,
		BlocksMap: blocksMap,
		KeysMap:   keysMap,
		logger:    mylogger.InitLogger("blockchain", id),
		mu:        sync.Mutex{},
	}
	return blockChain
}

func Block2Hash(block *Block) []byte {
	hash, _ := ComputeHash(block.Data)
	return hash
}

// key is a 20-byte ascii
func Hash2Key(hash []byte) string {
	var key []byte
	for i := 0; i < 20; i++ {
		key = append(key, uint8(97)+uint8(hash[i]%(26)))
	}
	return string(key)
}
func Block2Key(block *Block) string {
	return Hash2Key(Block2Hash(block))
}

// Thread-safe, append a block to the chain
// Internal have nothing to do with consensus, just append in a lock context
func (bc *BlockChain) AddBlockToChain(block *Block) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.Blocks = append(bc.Blocks, block)
	bc.KeysMap[block] = Block2Key(block)
	bc.BlocksMap[Block2Key(block)] = block
}

// Generate a Block: max rate is 20 blocks/s
func (bc *BlockChain) getBlock(seq uint64) *Block {
	//slow down
	time.Sleep(time.Duration(50) * time.Millisecond)
	data := make([]byte, bc.BlockSize)
	for i := uint64(0); i < bc.BlockSize; i++ {
		data[i] = byte(rand.Intn(256))
	}
	block := &Block{
		Seq:  seq,
		Data: data,
	}
	bc.logger.DPrintf("generate Block[%v] in seq %v at %v", Block2Key(block), block.Seq, time.Now().UnixNano())
	return block
}

// commit block to local BlockChain object
func (bc *BlockChain) commitBlock(block *Block) {
	bc.AddBlockToChain(block)
	bc.logger.DPrintf("commit Block[%v] in seq %v at %v", Block2Key(block), block.Seq, time.Now().UnixNano())
}
