package main

import "sync"

// KV represents a key value store
type KV struct {
	data map[string][]byte
	mu   sync.RWMutex
}

func NewKeyVal() *KV {
	return &KV{
		data: make(map[string][]byte),
	}
}

func (kv *KV) Set(key, val []byte) error {

	kv.mu.Lock()
	defer kv.mu.Unlock()

	kv.data[string(key)] = []byte(val)
	return nil
}

func (kv *KV) Get(key []byte) ([]byte, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	if _, ok := kv.data[string(key)]; !ok {
		return nil, false
	}

	return kv.data[string(key)], true
}
