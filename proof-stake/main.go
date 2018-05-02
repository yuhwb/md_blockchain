package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"fmt"
)

type Block struct {
	Index     uint64   `json:"index"`
	Timestamp int64 `json:"timestamp"`
	BPM       int    `json:"bpm"`
	PrevHash  string `json:"prevHash"`
	Hash      string `json:"hash"`
	Validator string `json:"validator"`
}

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
	record := strconv.FormatUint(b.Index , 16) + strconv.FormatInt(b.Timestamp , 16) + strconv.FormatInt(int64(b.BPM) , 16) + b.PrevHash + b.Validator
	fmt.Println(record)
	h =  CalculateHash(CalculateHash(record))
	return
}

func main() {

}
