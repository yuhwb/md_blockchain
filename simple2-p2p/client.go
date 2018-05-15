package main

import (
	"flag"
	"net"
	"log"
	"fmt"
	"bufio"
)

var server = flag.String("s" , "127.0.0.1:4009" ,"server address")
var local = flag.String("c", "127.0.0.1:4005" , "client address")
var dir = flag.String("d" , "D://11" , "data dir")

func main() {
	flag.Parse()
	addr , err := net.ResolveTCPAddr("tcp" , *local)
	if err != nil {
		log.Fatalln("error resolving client:" , err.Error())
		return 
	}

	conn , err := net.Dial("tcp" , *server)
	if err != nil {
		fmt.Println("error resolving server", err)
		return
	}

	b := make([]byte , 1024)
	n,err := conn.Read(b)
	s := string(b[:n])
	if err != nil {
		fmt.Println("error reading ", err)
	}
	if "hi" == s {
		fmt.Println("And that's fine")
	} else {
		fmt.Println("Error initializing", err)
	}
	conn.Write([]byte("I'm " + addr.String() + "\n"))

	conn.Write([]byte("list\n"))

	line , _ , err := bufio.NewReader(conn).ReadLine()
	fmt.Println(string(line))
}
