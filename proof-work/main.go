package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const ModifyNum = 200

var BlockChain []Block
var Difficulty = 5
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

type Message struct {
	BPM int `json:"bpm"`
}

// update Difficulty
func ModifyDifficulty() {
	index := len(BlockChain) - 1
	if (index+1)%ModifyNum == 0 {
		mutex.Lock()
		defer mutex.Unlock()
		if (index+1)%ModifyNum == 0 && (index+1)/ModifyNum > 1 {
			t1 := BlockChain[index].Timestamp - BlockChain[index-ModifyNum+1].Timestamp
			t2 := BlockChain[index-ModifyNum].Timestamp - BlockChain[index-ModifyNum*2+1].Timestamp
			Difficulty = int(t2 / t1)
		}
	}
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

	//计算工作量证明
	for i := 0; ; i++ {
		nonce := fmt.Sprintf("%x", i)
		newBlock.Nonce = nonce
		hash, _ := newBlock.GenerateHash()
		if !IsValidHash(hash, Difficulty) {
			fmt.Println(hash, "do more work")
			continue
		} else {
			fmt.Println(hash, "work done")
			newBlock.Hash = hash
			break
		}
	}
	return newBlock, nil
}

// 验证hash是否符合难度系数
func IsValidHash(hash string, difficulty int) bool {
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

func IsValidBlock(newBlock, oldBlock *Block) bool {
	if newBlock.Index != oldBlock.Index+1 {
		return false
	}

	if newBlock.PrevHash != oldBlock.Hash {
		return false
	}

	if !IsValidHash(newBlock.Hash, Difficulty) { //是否符合工作量证明
		return false
	}

	hash, err := newBlock.GenerateHash()
	if err != nil {
		return false
	}

	if hash != newBlock.Hash {
		return false
	}

	return true
}

func ReplaceChain(newBlocks []Block) {
	if len(newBlocks) > len(BlockChain) {
		mutex.Lock()
		defer mutex.Unlock()
		if len(newBlocks) > len(BlockChain) {
			BlockChain = newBlocks
		}
	}
}

func init() {
	genesisBlock := &Block{}
	genesisBlock.Index = 0
	genesisBlock.Timestamp = time.Now().UnixNano()
	genesisBlock.BPM = 0
	genesisBlock.Difficulty = Difficulty
	BlockChain = append(BlockChain, *genesisBlock)
}

func handleBlockChain(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		buf, err := json.MarshalIndent(BlockChain, "", "  ")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Header().Add("Content-Type", "application/json")
		w.Write(buf)
		return
	}
	if r.Method == http.MethodPost {
		var msg Message
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&msg); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		//调整难度系数
		ModifyDifficulty()
		//创建新的区块
		newBlock, err := GenerateBlock(&BlockChain[len(BlockChain)-1], msg.BPM)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if IsValidBlock(newBlock, &BlockChain[len(BlockChain)-1]) {
			newBlockChain := append(BlockChain, *newBlock)
			ReplaceChain(newBlockChain)
		}

		buf, _ := json.MarshalIndent(newBlock, "", "  ")
		w.WriteHeader(http.StatusCreated)
		w.Header().Add("Content-Type", "application/json")
		w.Write(buf)
	}
}

func main() {
	http.HandleFunc("/", handleBlockChain)
	server := &http.Server{
		Addr:           ":8080",
		Handler:        nil,
		//ReadTimeout:    10 * time.Second,
		//WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
