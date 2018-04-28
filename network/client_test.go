package main

import (
	"log"
	"net"
	"bytes"
	"encoding/binary"
	"strconv"
	"io"
	"math/rand"
	"bufio"
)

func main() {
	ch := make(chan int)
	num := 5
	for i:=0 ; i< num;i++ {
		go generateTcpClients()
	}
	<-ch
}

func generateTcpClients(){
	ch := make(chan bool)
	conn,err := net.Dial("tcp" , "localhost:9000")
	if err != nil {
		log.Fatal(err)
		return
	}
	defer conn.Close()

	go func() {
		scanner := bufio.NewScanner(conn)
		for scanner.Scan() {
			log.Println(scanner.Text())
		}
	}()

	go func(){
		for i := 0; i < 100; i++ {
			doWrite(conn , rand.Intn(100))
		}
	}()
	go func() {
		for i := 0; i < 100; i++ {
			doWrite(conn , rand.Intn(100))
		}
	}()

	ch <- true
}

func doWrite(conn net.Conn , n int)  {
	s := strconv.Itoa(n) + "\n"
	io.WriteString(conn , s)
}

func IntToBytes(n int) []byte{
	buf := bytes.NewBuffer(nil)
	binary.Write(buf , binary.BigEndian , n)
	return buf.Bytes()
}
