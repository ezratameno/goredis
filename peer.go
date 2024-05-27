package main

import (
	"errors"
	"io"
	"log"
	"net"

	"github.com/tidwall/resp"
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

// readLoop reads commands that come from the peer
func (p *Peer) readLoop() error {
	rd := resp.NewReader(p.conn)

	for {
		v, _, err := rd.ReadValue()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if v.Type() == resp.Array {

			if len(v.Array()) == 0 {
				return errors.New("empty array command not allowed, expects at lease one command")
			}

			commandName := v.Array()[0]

			switch commandName.String() {

			case CommandSet:

				if len(v.Array()) != 3 {
					return errors.New("set command expects 2 params")
				}
				cmd := SetCommand{
					key: v.Array()[1].Bytes(),
					val: v.Array()[2].Bytes(),
				}

				p.msgCh <- Message{
					cmd:  cmd,
					peer: p,
				}

			case CommandGet:
				if len(v.Array()) != 2 {
					return errors.New("get command expects 1 params")
				}

				cmd := GetCommand{
					key: v.Array()[1].Bytes(),
				}

				// Send the command to the server
				p.msgCh <- Message{
					cmd:  cmd,
					peer: p,
				}

			default:

			}

		}
	}

	return nil
}
