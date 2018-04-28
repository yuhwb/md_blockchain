package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"sync"
	"time"
)

const ModifyNum = 200

var BlockChain []Block
var Difficulty = 1
var mutex = &sync.Mutex{}

type Block struct {
	Index      int    `json:"index"`
	Timestamp  int64  `json:"timestamp"`
	BPM        int    `json:"bpm"`
	PrevHash   string `json:"prevHash"`
	Hash       string `json:"hash"`
	Difficulty int    `json:"difficulty"`
	Nonce      string `json:"nonce"`
}

//double hash256
func (b *Block) GenerateHash() (s string, err error) {
	if b == nil {
		return "", errors.New("nil pointer error")
	}
	record := strconv.Itoa(b.Index) + strconv.FormatInt(b.Timestamp, 10) + strconv.Itoa(b.BPM) + b.PrevHash + strconv.Itoa(b.Difficulty) + b.Nonce
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	h.Write(hashed)
	hashed = h.Sum(nil)
	s = hex.EncodeToString(hashed)
	return
}

// Difficulty 难度系数-- 每生成200块 更新难度系数
func GenerateBlock(oldBlock *Block, BPM int) (newBlock *Block, err error) {
	newBlock = &Block{}
	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = time.Now().UnixNano()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Difficulty = Difficulty
	// update difficulty

}

// update Difficulty
func ModifyDifficulty(index int) {
	if (index + 1) % ModifyNum == 0 {
		mutex.Lock()
		if (index + 1) % ModifyNum == 0 && (index + 1) / ModifyNum > 1 {
			t1 := BlockChain[index].Timestamp - BlockChain[index- ModifyNum + 1].Timestamp
			t2 := BlockChain[index- ModifyNum].Timestamp - BlockChain[index- ModifyNum*2 + 1].Timestamp
			Difficulty = int(t2/t1)
		}
		mutex.Unlock()
	}
}

func main() {

}
