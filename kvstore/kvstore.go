package kvstore

import "sync"

type KVStore struct {
	mu    sync.Mutex
	store map[string]string
}

func NewKVStore() *KVStore {
	return &KVStore{
		store: make(map[string]string),
	}
}

func (kv *KVStore) Put(key, value string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.store[key] = value
}

func (kv *KVStore) Append(key, value string) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	kv.store[key] += value
}

func (kv *KVStore) Get(key string) (string, bool) {
	kv.mu.Lock()
	defer kv.mu.Unlock()
	val, exists := kv.store[key]
	return val, exists
}