package main

import (
	"bufio"
	"log"
	"net"
	"os"
)

func main() {
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
			log.Print(string(data))
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		conn.Write(scanner.Bytes())
	}

}
