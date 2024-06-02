package main

import (
	"bytes"
	"fmt"
)

const (
	CommandSet   = "SET"
	CommandGet   = "GET"
	CommandHELLO = "hello"
)

// Redis commands are used to perform some operations on Redis server.
//
// To run commands on Redis server, you need a Redis client.
type Command interface {
}

// SetCommand set a key value in the store.
type SetCommand struct {
	key []byte
	val []byte
}

// GetCommand returns the value of the key from the store.
type GetCommand struct {
	key []byte
}

// HELLO always replies with a list of current server and connection properties,
// such as: versions, modules loaded, client ID, replication role and so forth.
type HelloCommand struct {
	value string
}

func respWriteMap(m map[string]string) []byte {
	buf := bytes.Buffer{}

	buf.WriteString("%" + fmt.Sprintf("%d\r\n", len(m)))

	for k, v := range m {
		buf.WriteString(fmt.Sprintf("+%s\r\n", k))
		buf.WriteString(fmt.Sprintf(":%s\r\n", v))
	}

	return buf.Bytes()
}
