package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestProtocol(t *testing.T) {

	raw := "*3\r\n$3\r\nSET\r\n$3\r\nfoo\r\n$3\r\nbar\r\n"
	cmd, err := parseCommand(raw)
	require.NoError(t, err)

	if _, ok := cmd.(SetCommand); !ok {
		require.Fail(t, "command should be a set command")
	}

	setCmd := cmd.(SetCommand)
	require.Equal(t, SetCommand{key: "foo", val: "bar"}, setCmd)
}
