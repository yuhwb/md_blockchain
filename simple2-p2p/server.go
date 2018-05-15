package main

import (
	"io"
	"log"
	"net"
	"regexp"
	"strings"
	"time"
	"bufio"
)

type TClient struct {
	local     string
	forserver string
	msg       string
}

type SFile struct {
	local string
	name  string
}

func MyReader(conn *net.TCPConn, clients chan TClient) {
	reader := bufio.NewReader(conn)
	for {
		l,_, err := reader.ReadLine()
		if err != nil {
			log.Fatalln("read error:",err.Error())
		}
		line := string(l)
		if regexp.MustCompile("^I'm").MatchString(line) {
			ladr, _ := net.ResolveTCPAddr("tcp", strings.Split(line, " ")[1])
			localAddr := ladr.String()
			log.Println(localAddr, conn.RemoteAddr().String())
			clients <- TClient{localAddr, conn.RemoteAddr().String(), "new"}
		} else {
			if len(line) > 0 {
				clients <- TClient{"", conn.RemoteAddr().String(), line}
				log.Println("data sent by client:", line)
				time.Sleep(5 * time.Millisecond)
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
	}
}

func ProcessConn(conn *net.TCPConn, clients chan TClient) {
	log.Println("connected.\n")
	io.WriteString(conn, "hi")
	go MyReader(conn, clients)
}

func ListenConnections(listener *net.TCPListener, connections chan *net.TCPConn, clients chan TClient) {
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Fatalln(err.Error())
			continue
		}
		conn.SetKeepAlive(true)
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		go ProcessConn(conn, clients)
		connections <- conn
	}
}

func main() {
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:4009")
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}

	connections := make(chan *net.TCPConn)
	clients := make(chan TClient)
	cMap := make(map[string]*net.TCPConn)
	fMap := make(map[string]string)

	go ListenConnections(listener, connections, clients)

	log.Println("waiting for connections.\n")
	for {
		select {
		case conn := <-connections:
			cMap[conn.RemoteAddr().String()] = conn
		case client := <-clients:
			if regexp.MustCompile("^new").MatchString(client.msg) {
				fMap[client.forserver] = client.local
			}
			if regexp.MustCompile("^list").MatchString(client.msg) {
				for key, value := range fMap {
					cMap[client.forserver].Write([]byte(key + "->" + value))
				}
				cMap[client.forserver].Write([]byte("\n"))
			}
		}
	}
}
