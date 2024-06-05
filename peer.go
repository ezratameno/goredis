package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/tidwall/resp"
)

// Peer represents the open connection to our server
type Peer struct {
	conn  net.Conn
	msgCh chan Message
	delCh chan *Peer
}

func NewPeer(conn net.Conn, msgCh chan Message, delCh chan *Peer) *Peer {
	return &Peer{
		conn:  conn,
		msgCh: msgCh,
		delCh: delCh,
	}
}

func (p *Peer) Send(msg []byte) (int, error) {
	return p.conn.Write(msg)
}

// readLoop reads and parse the commands that come from the peer, and returns the command to the server.
func (p *Peer) readLoop() error {
	rd := resp.NewReader(p.conn)

	for {
		v, _, err := rd.ReadValue()

		// Connection closed
		if err == io.EOF {
			p.delCh <- p
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

			var cmd Command
			switch commandName.String() {

			case CommandSet:

				if len(v.Array()) != 3 {
					return errors.New("set command expects 2 params")
				}
				cmd = SetCommand{
					key: v.Array()[1].Bytes(),
					val: v.Array()[2].Bytes(),
				}

			case CommandGet:
				if len(v.Array()) != 2 {
					return errors.New("get command expects 1 params")
				}

				cmd = GetCommand{
					key: v.Array()[1].Bytes(),
				}

			case CommandHELLO:

				if len(v.Array()) != 2 {
					return errors.New("hello command expects 1 params")
				}

				cmd = HelloCommand{
					value: v.Array()[1].String(),
				}

			default:
				fmt.Printf("got unknown command => %v\n", v.Array())
			}

			// Send the command to the server
			p.msgCh <- Message{
				cmd:  cmd,
				peer: p,
			}
		}
	}

	return nil
}
