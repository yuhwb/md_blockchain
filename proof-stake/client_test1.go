package main

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"net"
	"strconv"
	"time"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9000")
	if err != nil {
		fmt.Println(err)
		return
	}
	r1 := 100
	r2 := 200

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		s := scanner.Text()
		fmt.Print(s)
		switch s {
		case "Enter token balance:":
			s := strconv.Itoa(rand.Intn(r1))
			fmt.Println(s)
			io.WriteString(conn, s+"\n")
		case "Enter a new BPM:":
			time.Sleep(time.Duration(rand.Intn(2000)) * time.Millisecond)
			s := strconv.Itoa(rand.Intn(r2))
			fmt.Println(s)
			io.WriteString(conn, s+"\n")
		default:
			fmt.Println()
		}
	}
}
