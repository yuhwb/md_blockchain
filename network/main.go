package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"
	"net"
	"log"
	"os"
	"bufio"
	"io"
	"encoding/json"
	"sync"
)

const (
	PORT = 9000
)

var BlockChain []Block
var bcServer chan []Block //存放链的切片的通道
var mutex = &sync.Mutex{}

type Block struct {
	Index     int    `json:"index"`
	Timestamp int64  `json:"timestamp"`
	BPM       int    `json:"bpm"`
	PrevHash   string `json:"prevHash"`
	Hash      string `json:"hash"`
}


func (b *Block) GenerateHash() (string, error) {
	if b == nil {
		return "", errors.New("nil pointer error")
	}
	record := strconv.Itoa(b.Index) + strconv.FormatInt(b.Timestamp, 10) + strconv.Itoa(b.BPM) + b.PrevHash
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	fmt.Println(string(hashed))
	return hex.EncodeToString(hashed), nil
}

func GenerateBlock(oldBlock *Block , BPM int) (newBlock *Block, err error) {
	newBlock = &Block{}
	newBlock.Index = oldBlock.Index + 1
	newBlock.Timestamp = time.Now().UnixNano()
	newBlock.BPM = BPM
	newBlock.PrevHash = oldBlock.Hash
	hashed,err :=newBlock.GenerateHash()
	if err != nil {
		return nil,err
	}
	newBlock.Hash = hashed
	return
}

func IsValidBlock(newBlock,oldBlock *Block) bool{
	if newBlock.Index != oldBlock.Index + 1 {
		return false
	}
	if newBlock.PrevHash != oldBlock.Hash {
		return false
	}
	hashed ,err := newBlock.GenerateHash()
	if err != nil {
		return false
	}
	if hashed != newBlock.Hash {
		return false
	}
	return true
}

func ReplaceChain(newBlocks []Block)  {
	mutex.Lock()
	defer mutex.Unlock()
	if len(newBlocks) > len(BlockChain) {
		BlockChain = newBlocks
	}
}

//创建创世区块并添加到区块链中
func init() {
	genesisBlock := Block{}
	genesisBlock.Index = 0
	genesisBlock.Timestamp = time.Now().UnixNano()
	genesisBlock.BPM = 0
	BlockChain = append(BlockChain , genesisBlock)
}

/*func ScanLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '.'); i >= 0 {
		// We have a full newline-terminated line.
		return i + 1, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

// dropCR drops a terminal \r from the data.
func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '.' {
		return data[0 : len(data)-1]
	}
	return data
}*/

//编写链接请求处理
func handleConn(conn net.Conn)  {
	defer conn.Close()
	io.WriteString(conn, "Enter a new BPM:")

	// 接收数据，新生成区块加入到新链中
	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			fmt.Printf("receive a BPM:%v \n" , scanner.Text())
			bpm,err := strconv.Atoi(scanner.Text())
			if err != nil {
				log.Printf("%v not a number:%v" , scanner.Text() , err)
				continue
			}
			newBlock , err :=GenerateBlock(&BlockChain[len(BlockChain) - 1] , bpm)
			if err != nil {
				log.Println(err)
				continue
			}
			if IsValidBlock(newBlock , &BlockChain[len(BlockChain) - 1]) {
				newBlockChain := append(BlockChain , *newBlock)
				ReplaceChain(newBlockChain)
			}
			bcServer <- BlockChain
			io.WriteString(conn , "Enter a new BPM:")
		}
	}()

	//每30秒向全网广播数据
	go func() {
		for  {
			time.Sleep(30*time.Second)
			mutex.Lock()
			output , err := json.MarshalIndent(BlockChain ,"" , "\t")
			if err != nil {
				log.Fatalln(err)
				continue
			}
			mutex.Unlock()
			conn.Write(output)
		}
	}()

	for _ =range bcServer{
		printPrettyJson()
	}
}

func main() {
	bcServer = make(chan []Block)

	server,err := net.Listen("tcp" , ":"+strconv.Itoa(PORT))
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
	defer server.Close()

	for  {
		conn,err := server.Accept() //阻塞
		if err != nil {
			log.Fatal(err)
			continue
		}
		go handleConn(conn) //开启协程处理
	}
}

func printPrettyJson() string {
	log.Println("1213")
	data ,err :=json.MarshalIndent(BlockChain , "" ,"\t")
	if err != nil {
		log.Fatalln(err)
		return ""
	}
	s := string(data)
	log.Println(s)
	return s
}
