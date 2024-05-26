package main

import (
	"net"
)

// Peer represents the open connection to our server
type Peer struct {
	conn  net.Conn
	msgCh chan Message
}

func NewPeer(conn net.Conn, msgCh chan Message) *Peer {
	return &Peer{
		conn:  conn,
		msgCh: msgCh,
	}
}

func (p *Peer) Send(msg []byte) (int, error) {
	return p.conn.Write(msg)
}

func (p *Peer) readLoop() error {

	for {
		buf := make([]byte, 4096)

		n, err := p.conn.Read(buf)
		if err != nil {
			return err
		}

		msgBuf := make([]byte, n)
		copy(msgBuf, buf)

		// send the message to the server
		p.msgCh <- Message{
			data: buf,
			peer: p,
		}
	}
}
