package main

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestOfficialRedisClient(t *testing.T) {

	listernAddr := "5001"
	server := NewServer(Config{
		ListenAddr: fmt.Sprintf(":%s", listernAddr),
	})

	go func() {
		log.Fatal(server.Start())
	}()

	time.Sleep(500 * time.Millisecond)

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("localhost:%s", listernAddr),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	defer rdb.Close()

	key := "foo"
	val := "bar"
	ctx := context.Background()
	err := rdb.Set(ctx, key, val, 0).Err()
	require.NoError(t, err, "set error")

	newVal, err := rdb.Get(ctx, key).Result()
	require.NoError(t, err, "get error")

	require.Equal(t, val, newVal)

}

func TestFooBar(t *testing.T) {
	in := map[string]string{
		"server":  "redis",
		"version": "6.0",
		"proto":   "3",
		"mode":    "standalone",
		"role":    "master",
	}
	out := respWriteMap(in)

	fmt.Println(string(out))

}
