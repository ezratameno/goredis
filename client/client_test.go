package client

import (
	"context"
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewClient1(t *testing.T) {
	client, err := New("localhost:5001")
	require.NoError(t, err)
	defer client.Close()

	err = client.Set(context.Background(), "foo", "1")
	require.NoError(t, err)

	value, err := client.Get(context.Background(), "foo")
	require.NoError(t, err)

	n, err := strconv.Atoi(value)
	require.NoError(t, err)
	fmt.Println(n)
	fmt.Println("value", value)

}

func TestNewClient(t *testing.T) {
	client, err := New("localhost:5001")
	require.NoError(t, err)
	defer client.Close()

	for i := 0; i < 10; i++ {

		err = client.Set(context.Background(), fmt.Sprintf("foo_%d", i), fmt.Sprintf("bar_%d", i))
		require.NoError(t, err)

		value, err := client.Get(context.Background(), fmt.Sprintf("foo_%d", i))
		require.NoError(t, err)

		fmt.Println("value", value)
	}
}
