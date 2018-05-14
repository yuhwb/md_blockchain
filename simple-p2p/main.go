package main

import (
	"flag"
	"fmt"
)

var (
	peers int
)

func init()  {
	flag.IntVar(&peers , "peers" , 1000 , "peers's numbers,if it's 0 will allow an unlimited amount of peers to connect to the server.")
}

func main() {
	flag.Parse()
	fmt.Println("peers" , peers)

	peer := NewPeer(1000)
	peer.Start()
}
