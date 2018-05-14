package main

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"io"
	"log"
	"math"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Peer struct {
	Hash     hash.Hash
	MaxPeers int
	Port     string
	Host     string
	PeerID   string
	PeerLock *sync.Mutex
	Peers    map[string]net.Addr
	shutdown bool
	Handler *Handler
}

func NewPeer(maxPeers int) *Peer {
	p := &Peer{
		Hash:     sha256.New(),
		MaxPeers: maxPeers,
		Port:     "8888",
		Peers:    make(map[string]net.Addr),
		shutdown: false,
	}

	if maxPeers == 0 {
		p.MaxPeers = math.MaxInt64
	}

	p.init()

	p.Handler = &Handler{p}

	return p
}

func (p *Peer) init() {
	ifaces, _ := net.Interfaces() //返回该系统的网络接口列表
	for _, i := range ifaces {
		addrs, _ := i.Addrs() //返回网络接口的一个或多个接口地址
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			p.Host = ip.String()
			io.WriteString(p.Hash, p.Host+":"+p.Port)
			p.PeerID = hex.EncodeToString(p.Hash.Sum(nil))
		}
	}
}

func (p *Peer) Start() {
	l, err := net.Listen("tcp", ":"+p.Port)
	if err != nil {
		log.Fatalln(err.Error())
		return
	}
	defer l.Close()

	go p.listenForShutdown()
	go p.CheckLivePeers()

	for !p.shutdown {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln(err.Error())
			continue
		}
		go p.Handler.HandleConn(conn)
	}

	log.Printf("Peer [%s] is shutting down.\n" , p.PeerID)
}

func (p *Peer) ShutDown() {
	p.shutdown = true
}

func (p *Peer) listenForShutdown() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	go func() {
		msg := "Starting shutdown."
		for s := range sig {
			log.Printf("Received signal:%v.%s\n", s, msg)
			msg = "Shutdown in progress."
			p.ShutDown()
			return
		}
	}()
}

func (p *Peer) CheckLivePeers() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for t := range ticker.C {
		fmt.Println(t.String())
		if p.shutdown {
			return
		}
		log.Println("Checking if known peers are live...")
		for _, addr := range p.Peers {
			peerID := addr.String() + ":" + p.Port
			if !p.Handler.Ping(peerID) {
				log.Println("Unable to ping:", peerID)
				log.Println("Removeing peer:", peerID)
				p.RemovePeer(peerID)
			}
		}
	}
}

func (p *Peer) AddPeer(peerID string, addr net.Addr) error {
	p.PeerLock.Lock()
	defer p.PeerLock.Unlock()
	if len(p.Peers) >= p.MaxPeers {
		return errors.New("Max peer limit has been reached for peer :" + peerID)
	}
	if _,ok := p.Peers[peerID] ; !ok {
		p.Peers[peerID] = addr
		log.Println("Added" , peerID , "to list of peers.")
		return nil
	}
	log.Printf("Peer [%s] already in list of known peers.\n" , peerID)
	return nil
}

func (p *Peer) RemovePeer(peerID string) error {
	p.PeerLock.Lock()
	defer p.PeerLock.Unlock()
	if _, ok := p.Peers[peerID]; ok {
		delete(p.Peers, peerID)
		log.Println("Successfully removed peer:", peerID)
		return nil
	}
	log.Printf("Unable to remove peer [%s] from list of known peers.", peerID)
	return fmt.Errorf("Peer [%s] not found in known peers.", peerID)
}
