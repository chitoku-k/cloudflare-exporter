package service

import (
	"context"
	"net"
	"time"
)

type Probe interface {
	Collect(ctx context.Context, url string) (Trace, error)
}

type Trace struct {
	FL          string
	H           string
	IP          net.IP
	TS          time.Time
	VisitScheme string
	UAG         string
	Colo        string
	Sliver      string
	Loc         string
	TLS         string
	SNI         string
	Warp        string
	Gateway     string
	RBI         string
	KEX         string
}
