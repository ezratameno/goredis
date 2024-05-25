package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"time"

	"github.com/ezratameno/goredis/client"
)

func main() {
	err := run()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

const (
	defaultListenAddr = ":5001"
)

type Config struct {
	ListenAddr string
}

type Server struct {
	Config
	ln net.Listener

	// open connections to our server
	peers     map[*Peer]bool
	addPeerCh chan *Peer

	quitCh chan struct{}

	msgCh chan []byte

	// key value store
	kv *KV
}

func NewServer(cfg Config) *Server {

	if len(cfg.ListenAddr) == 0 {
		cfg.ListenAddr = defaultListenAddr
	}

	return &Server{
		Config:    cfg,
		peers:     make(map[*Peer]bool),
		addPeerCh: make(chan *Peer),
		quitCh:    make(chan struct{}),
		msgCh:     make(chan []byte),
		kv:        NewKeyVal(),
	}
}

// Start starts to listen on the port
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.ListenAddr)
	if err != nil {
		return err
	}

	s.ln = ln

	go s.loop()

	slog.Info("server running", "listenAddr", s.ListenAddr)

	return s.acceptLoop()

}

// acceptLoop listens to new connections to the server.
func (s *Server) acceptLoop() error {

	for {

		// waits for the next connection to our server

		conn, err := s.ln.Accept()
		if err != nil {
			slog.Error("accept error", "err", err)
			continue
		}

		go s.handleConn(conn)

	}
}

func (s *Server) loop() {

	for {
		select {
		case <-s.quitCh:
			return

		// Add the peer
		case peer := <-s.addPeerCh:
			s.peers[peer] = true

		// Read from the peers
		case rawMsg := <-s.msgCh:

			err := s.handleRawMessage(rawMsg)
			if err != nil {
				slog.Error("raw message error", "err", err)
			}
		}
	}

}

// handleRawMessage handles the raw message from the peer.
func (s *Server) handleRawMessage(rawMsg []byte) error {

	cmd, err := parseCommand(string(rawMsg))
	if err != nil {
		return err
	}

	// check the type of the command
	switch v := cmd.(type) {
	case SetCommand:
		slog.Info("somebody want to set a key to the hash table", "key", v.key, "val", v.val)
		return s.kv.Set(v.key, v.val)

	case GetCommand:

	}

	return nil
}

// handleConn creates a peer from the connection and read from him.
func (s *Server) handleConn(conn net.Conn) {

	// Add the peer
	peer := NewPeer(conn, s.msgCh)
	s.addPeerCh <- peer

	slog.Info("new peer connected", "remoteAddr", peer.conn.RemoteAddr())

	// read from the new peer
	err := peer.readLoop()
	if err != nil {
		slog.Error("peer read error", "err", err)
	}
}

func run() error {
	server := NewServer(Config{})

	go func() {

		err := server.Start()
		if err != nil {
			log.Fatal(err)
		}
	}()

	// wait for the server to boot
	time.Sleep(time.Second * 2)

	for i := 0; i < 10; i++ {
		client, err := client.New("localhost:5001")
		if err != nil {
			return err
		}
		err = client.Set(context.Background(), fmt.Sprintf("foo_%d", i), fmt.Sprintf("bar_%d", i))
		if err != nil {
			return err
		}
	}

	time.Sleep(2 * time.Second)

	fmt.Printf("%+v\n", server.kv.data)

	return nil
}
