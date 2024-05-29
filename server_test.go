package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/ezratameno/goredis/client"
	"github.com/stretchr/testify/require"
)

func TestServerWithMultiClients(t *testing.T) {

	server := NewServer(Config{})

	go func() {

		log.Fatal(server.Start())
	}()

	time.Sleep(1 * time.Second)

	nClients := 10
	var wg sync.WaitGroup
	for i := 0; i < nClients; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client, err := client.New("localhost:5001")
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

	time.Sleep(500 * time.Microsecond)

	if len(server.peers) != 0 {
		t.Fatalf("expected 0 peers but got %d ", len(server.peers))
	}

}
