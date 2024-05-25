package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/tidwall/resp"
)

const (
	CommandSet = "SET"
)

// Redis commands are used to perform some operations on Redis server.
//
// To run commands on Redis server, you need a Redis client.
type Command interface {
}

// SetCommand implements the redis set command
type SetCommand struct {
	key string
	val string
}

func parseCommand(raw string) (Command, error) {

	rd := resp.NewReader(bytes.NewBufferString(raw))
	var cmd Command

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
				return nil, errors.New("empty array command not allowed, expects at lease one command")
			}

			commandName := v.Array()[0]

			switch commandName.String() {

			case CommandSet:

				if len(v.Array()) != 3 {
					return nil, errors.New("set command expects 2 params")
				}
				cmd = SetCommand{
					key: v.Array()[1].String(),
					val: v.Array()[2].String(),
				}

				return cmd, nil
			default:

			}

		}
	}

	return cmd, fmt.Errorf("invalid or unknown command received: %s", raw)

}
