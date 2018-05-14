package main

import (
	"log"
	"net"
	"io"
	"bufio"
	"encoding/hex"
	"fmt"
)

var MaxBuffer = 1024

type Handler struct {
	Server *Peer
}

func (h *Handler) HandleConn(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, MaxBuffer)

	n, err := conn.Read(buffer)

	request := string(buffer[:n])

	if err != nil {
		log.Printf("Error reading message from connection [%s]: %s\n", conn.RemoteAddr().String(), err.Error())
	} else {
		log.Printf("Request from %s: %s\n", conn.RemoteAddr().String(), request)
	}

	response := "request type not recognized"
	switch request {
	case "PING":response = h.HealthCheck()
	case "REGISTER":response =h.RegisterHandler(conn)
	case "ECHO":response =h.EchoHandler(request)
	default:
	}

	conn.Write([]byte(response))
}

func (h *Handler) Ping(peerID string) bool {
	response := false

	conn,err := net.Dial("tcp" , peerID)
	if err != nil {
		log.Printf("Error dialing %s:%s\n" ,peerID , err.Error())
	} else {
		io.WriteString(conn , "PING")
		status , err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Printf("Error receiving PING from %s:%s\n" , peerID , err.Error())
			h.Server.RemovePeer(peerID)
		}else {
			log.Printf("Successful response from %s:%s\n" , peerID , status)
			response = true
		}
	}
	return response
}

func (h *Handler) HealthCheck() string {
	return "PONG"
}

func (h *Handler) RegisterHandler(conn net.Conn) string {
	addr := conn.RemoteAddr()
	io.WriteString(h.Server.Hash , addr.String() + ":8888")
	peerId := hex.EncodeToString(h.Server.Hash.Sum(nil))

	err := h.Server.AddPeer(peerId , addr)

	if err != nil {
		return fmt.Sprintf("[ERROR] Unable to add to this peer: %s", err.Error())
	} else {
		return "Successfully added Peer!"
	}
}

func (h *Handler) EchoHandler(message string) string {
	return message
}
