package main

import "time"

type Config struct {
	NodeIDs           []int       
	ElectionTimeout   time.Duration
	HeartbeatInterval time.Duration 

func DefaultConfig() *Config {
	return &Config{
		NodeIDs:           []int{1, 2, 3},        
		ElectionTimeout:   time.Millisecond * 150,
		HeartbeatInterval: time.Millisecond * 50,
	}
}
