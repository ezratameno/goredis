package client

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

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
