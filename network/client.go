package main

import (
	"log"
	"net"
	"os"
	"bytes"
	"encoding/binary"
	"strconv"
	"io"
	"fmt"
)

func main() {
	ch := make(chan int)
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		log.Fatal(1,err)
		os.Exit(1)
	}
	defer conn.Close()


	go func() {
		for {
			var buf [200]byte
			data, err := conn.Read(buf[:])
			if err != nil {
				log.Printf("error:%v", err)
			}
			fmt.Println(string(data))
		}
	}()

	for i := 0; i < 100; i++ {
		go doWrite(conn , 1)
	}
	for i := 0; i < 100; i++ {
		go doWrite(conn , 2)
	}
	for i := 0; i < 100; i++ {
		go doWrite(conn , 3)
	}
	for i := 0; i < 100; i++ {
		go doWrite(conn , 4)
	}
	for i := 0; i < 100; i++ {
		go doWrite(conn , 5)
	}

	<-ch

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
