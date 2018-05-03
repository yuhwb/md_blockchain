package main

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type Block struct {
	Index     uint64 `json:"index"`
	Timestamp int64  `json:"timestamp"`
	BPM       int    `json:"bpm"`
	PrevHash  string `json:"prevHash"`
	Hash      string `json:"hash"`
	Validator string `json:"validator"`
}

var (
	BlockChain      []Block
	tempBlocks      []Block
	mutex           = &sync.Mutex{}
	validators      = make(map[string]int)
	candidateBlocks = make(chan Block)
	announcements   = make(chan string)
)

func CalculateHash(s string) string {
	h := sha256.New()
	h.Write([]byte(s))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func (b *Block) CalculateBlockHash() (h string, err error) {
	if b == nil {
		return "", errors.New("nil pointer error")
	}
	record := strconv.FormatUint(b.Index, 16) + strconv.FormatInt(b.Timestamp, 16) + strconv.FormatInt(int64(b.BPM), 16) + b.PrevHash + b.Validator
	h = CalculateHash(CalculateHash(record))
	return
}

func GenerateBlock(oldBlock *Block, BPM int, address string) (newBlock *Block, err error) {
	newBlock = &Block{}
	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = time.Now().UnixNano()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	newBlock.Validator = address
	h, err := newBlock.CalculateBlockHash()
	if err != nil {
		return nil, err
	}
	newBlock.Hash = h
	return
}

// 验证新区快是否符合要求
func IsValidBlock(newBlock, oldBlock *Block) bool {
	if newBlock == nil || oldBlock == nil {
		return false
	}

	if newBlock.Index != oldBlock.Index+1 {
		return false
	}

	if newBlock.PrevHash != oldBlock.Hash {
		return false
	}

	h, err := newBlock.CalculateBlockHash()
	if err != nil {
		return false
	}

	if h != newBlock.Hash {
		return false
	}

	return true
}

//初始化  生成创世区块
func init() {
	genesisBlock := &Block{}
	genesisBlock.Index = 0
	genesisBlock.Timestamp = time.Now().UnixNano()
	genesisBlock.BPM = 0
	BlockChain = append(BlockChain, *genesisBlock)
}

// 挑选获胜者
func pickWinner() {
	time.Sleep(3 * time.Second)
	mutex.Lock()
	defer mutex.Unlock()
	lotteryPool := make([]string, 0)
	if len(tempBlocks) > 0 {
	OUTER:
		for _, block := range tempBlocks {
			for _, node := range lotteryPool {
				if node == block.Validator {
					continue OUTER
				}
			}
			v, ok := validators[block.Validator]
			if ok {
				for i := 0; i < v; i++ {
					lotteryPool = append(lotteryPool, block.Validator)
				}
			}
		}

		s := rand.NewSource(time.Now().UnixNano())
		r := rand.New(s)
		lotteryWinner := lotteryPool[r.Intn(len(lotteryPool))]

		for _, block := range tempBlocks {
			if block.Validator == lotteryWinner {
				BlockChain = append(BlockChain, block)
				for _ = range validators {
					s := "winning validator:" + lotteryWinner + "\n"
					announcements <- s
				}
				break
			}
		}
		tempBlocks = []Block{}
	}

}

func handleConn(conn net.Conn) {
	defer conn.Close()

	go func() {
		for {
			s := <-announcements
			io.WriteString(conn, s)
		}
	}()

	var address string

	io.WriteString(conn, "Enter token balance:\n")
	balanceScanner := bufio.NewScanner(conn)
	for balanceScanner.Scan() {
		balance, err := strconv.Atoi(balanceScanner.Text())
		if err != nil {
			log.Fatalf("%v not a number:%v", balanceScanner.Text(), err.Error())
			return
		}
		t := time.Now()
		address = CalculateHash(t.String())
		validators[address] = balance
		break
	}

	go func() {
		io.WriteString(conn, "Enter a new BPM:\n")
		bpmScanner := bufio.NewScanner(conn)
		for bpmScanner.Scan() {
			bpm, err := strconv.Atoi(bpmScanner.Text())
			if err != nil {
				log.Fatalf("%v not a number:%v", bpmScanner.Text(), err.Error())
				delete(validators, address)
				conn.Close()
			}
			mutex.Lock()
			oldBlock := &BlockChain[len(BlockChain)-1]
			mutex.Unlock()

			newBlock, err := GenerateBlock(oldBlock, bpm, address)
			if err != nil {
				log.Fatalln(err.Error())
				continue
			}
			// 验证区块是否合法，若合法则向通道写入数据
			if IsValidBlock(newBlock, oldBlock) {
				candidateBlocks <- *newBlock
			}
			io.WriteString(conn, "Enter a new BPM:\n")
		}
	}()

	// 广播区块链
	for {
		time.Sleep(30 * time.Second)
		mutex.Lock()
		output, err := json.MarshalIndent(BlockChain, "", "  ")
		mutex.Unlock()
		if err != nil {
			log.Fatalln(err.Error())
			continue
		}
		s := string(output)
		log.Println("\n" + s)
		io.WriteString(conn, s+"\n")
	}
}

func main() {
	l, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalln(err.Error())
		os.Exit(1)
	}
	defer l.Close()

	go func() {
		for candidateBlock := range candidateBlocks {
			mutex.Lock()
			tempBlocks = append(tempBlocks, candidateBlock)
			mutex.Unlock()
		}
	}()

	go func() {
		for {
			pickWinner()
		}
	}()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln(conn.RemoteAddr(), err)
			continue
		}
		go handleConn(conn) //开启一个goroutine
	}
}
