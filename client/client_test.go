package client

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewClients(t *testing.T) {

	nClients := 10

	var wg sync.WaitGroup
	for i := 0; i < nClients; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := New("localhost:5001")
			require.NoError(t, err)

			defer client.Close()

			key := fmt.Sprintf("client_%d", i)
			value := fmt.Sprintf("client_bar_%d", i)
			err = client.Set(context.Background(), key, value)
			require.NoError(t, err)

			val, err := client.Get(context.Background(), key)
			require.NoError(t, err)

			fmt.Printf("client %d got this value back => %s\n", i, val)
		}()

	}

	wg.Wait()

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
