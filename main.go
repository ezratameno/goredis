package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"reflect"

	"github.com/tidwall/resp"
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
	delPeerCh chan *Peer

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
		delPeerCh: make(chan *Peer),
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
			slog.Info("new peer connected", "remoteAddr", peer.conn.RemoteAddr())

			s.peers[peer] = true

		case peer := <-s.delPeerCh:
			slog.Info("peer disconnected", "remoteAddr", peer.conn.RemoteAddr())

			delete(s.peers, peer)

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

	slog.Info("got message from client", "type", reflect.TypeOf(msg.cmd))
	// check the type of the command
	switch v := msg.cmd.(type) {
	case SetCommand:
		slog.Info("somebody want to set a key to the hash table", "key", v.key, "val", v.val)
		err := s.kv.Set(v.key, v.val)

		if err != nil {
			return err
		}

		buf := &bytes.Buffer{}
		rw := resp.NewWriter(buf)
		rw.WriteString("OK")

		_, err = msg.peer.Send(buf.Bytes())
		if err != nil {
			return fmt.Errorf("peer send error: %w", err)
		}

	case GetCommand:
		slog.Info("somebody want to get a key from the hash table", "key", v.key)
		val, ok := s.kv.Get(v.key)
		if !ok {
			return fmt.Errorf("key %s not found", v.key)
		}

		buf := &bytes.Buffer{}
		rw := resp.NewWriter(buf)
		rw.WriteBytes(val)

		// Send the value of the key to the connection
		_, err := msg.peer.Send(buf.Bytes())
		if err != nil {
			return fmt.Errorf("peer send error: %w", err)
		}

	case HelloCommand:
		fmt.Println("this is the hello command from the client:", v.value)

		// Send the server spec to the client
		spec := map[string]string{
			"server":  "redis",
			"version": "6.0",
			"proto":   "3",
			"mode":    "standalone",
			"role":    "master",
		}
		_, err := msg.peer.Send((respWriteMap(spec)))
		if err != nil {
			return fmt.Errorf("peer send error: %w", err)
		}

	case ClientCommand:
		buf := &bytes.Buffer{}
		rw := resp.NewWriter(buf)
		rw.WriteString("OK")

		_, err := msg.peer.Send(buf.Bytes())
		if err != nil {
			return err
		}
		fmt.Printf("client command: %+v\n", v)

	default:
		fmt.Printf("unknown command => %+v\n", msg.cmd)
	}

	return nil
}

// handleConn creates a peer from the connection and read from him.
func (s *Server) handleConn(conn net.Conn) {

	// Add the peer
	peer := NewPeer(conn, s.msgCh, s.delPeerCh)
	s.addPeerCh <- peer

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
