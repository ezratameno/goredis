package main

import (
	"net"
)

// Peer represents the open connection to our server
type Peer struct {
	conn  net.Conn
	msgCh chan []byte
}

func NewPeer(conn net.Conn, msgCh chan []byte) *Peer {
	return &Peer{
		conn:  conn,
		msgCh: msgCh,
	}
}

func (p *Peer) readLoop() error {

	buf := make([]byte, 1024)
	for {
		n, err := p.conn.Read(buf)
		if err != nil {
			return err
		}

		msgBuf := make([]byte, n)
		copy(msgBuf, buf)

		// send the message to the server

		p.msgCh <- msgBuf
	}
}
