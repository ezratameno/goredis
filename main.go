package main

import (
	"fmt"
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

type Server struct {
	Config
	ln net.Listener

	// open connections to our server
	peers     map[*Peer]bool
	addPeerCh chan *Peer

	quitCh chan struct{}
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

		}
	}

}

func (s *Server) handleConn(conn net.Conn) {

	// Add the peer
	peer := NewPeer(conn)
	s.addPeerCh <- peer

	slog.Info("new peer connected", "remoteAddr", peer.conn.RemoteAddr())

	// read for the new peer
	err := peer.readLoop()
	if err != nil {
		slog.Error("peer read error", "err", err)
	}
}

func run() error {

	server := NewServer(Config{})

	err := server.Start()
	if err != nil {
		return err
	}
	return nil
}
