package service

import (
	"time"
)

type LoadBalancer interface {
	Collect(poolName string) ([]Pool, error)
}

type Pool struct {
	Name       string
	PopHealths []PopHealth
}

type PopHealth struct {
	Region  string
	Healthy bool
	Origins []Origin
}

type Origin struct {
	Address      string
	Healthy      bool
	RTT          time.Duration
	ResponseCode int
}
