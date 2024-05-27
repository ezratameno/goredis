package main

import (
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
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

type Message struct {
	cmd  Command
	peer *Peer
}

type Server struct {
	Config
	ln net.Listener

	// open connections to our server
	peers     map[*Peer]bool
	addPeerCh chan *Peer

	quitCh chan struct{}

	msgCh chan Message

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
		msgCh:     make(chan Message),
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

			err := s.handleMessage(rawMsg)
			if err != nil {
				slog.Error("raw message error", "err", err)
			}
		}
	}

}

// handleRawMessage handles the raw message from the peer.
func (s *Server) handleMessage(msg Message) error {

	// check the type of the command
	switch v := msg.cmd.(type) {
	case SetCommand:
		slog.Info("somebody want to set a key to the hash table", "key", v.key, "val", v.val)
		return s.kv.Set(v.key, v.val)

	case GetCommand:
		slog.Info("somebody want to get a key from the hash table", "key", v.key)
		val, ok := s.kv.Get(v.key)
		if !ok {
			return fmt.Errorf("key %s not found", v.key)
		}

		// Send the value of the key to the connection
		_, err := msg.peer.Send(val)
		if err != nil {
			slog.Error("peer send error", "err", err)
			break
		}

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

	listenAddr := flag.String("listenAddr", defaultListenAddr, "listen address of the goredis server")
	flag.Parse()
	server := NewServer(Config{
		ListenAddr: *listenAddr,
	})

	err := server.Start()
	if err != nil {
		log.Fatal(err)
	}

	return nil
}
