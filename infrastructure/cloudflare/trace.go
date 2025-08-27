package cloudflare

import (
	"bufio"
	"context"
	"fmt"
	"net"
	"net/http"
	neturl "net/url"
	"strconv"
	"strings"
	"time"

	"github.com/chitoku-k/cloudflare-exporter/service"
)

type probeService struct {
	Client *http.Client
}

func NewProbeService(client *http.Client) service.Probe {
	if client == nil {
		client = http.DefaultClient
	}
	return &probeService{
		Client: client,
	}
}

func (s *probeService) Collect(ctx context.Context, url string) (service.Trace, error) {
	u, err := neturl.JoinPath(url, "/cdn-cgi/trace")
	if err != nil {
		return service.Trace{}, fmt.Errorf("failed to parse URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return service.Trace{}, fmt.Errorf("failed to construct a request: %w", err)
	}

	res, err := s.Client.Do(req)
	if err != nil {
		return service.Trace{}, fmt.Errorf("failed to request: %w", err)
	}
	defer func() {
		_ = res.Body.Close()
	}()

	if res.StatusCode != http.StatusOK {
		return service.Trace{}, fmt.Errorf("failed to trace: status: %d", res.StatusCode)
	}

	var trace service.Trace
	fields := map[string]any{
		"fl":           &trace.FL,
		"h":            &trace.H,
		"ip":           &trace.IP,
		"ts":           &trace.TS,
		"visit_scheme": &trace.VisitScheme,
		"uag":          &trace.UAG,
		"colo":         &trace.Colo,
		"sliver":       &trace.Sliver,
		"loc":          &trace.Loc,
		"tls":          &trace.TLS,
		"sni":          &trace.SNI,
		"warp":         &trace.Warp,
		"gateway":      &trace.Gateway,
		"rbi":          &trace.RBI,
		"kex":          &trace.KEX,
	}

	scanner := bufio.NewScanner(res.Body)
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if !ok {
			return service.Trace{}, fmt.Errorf("failed to parse: %s", scanner.Text())
		}

		field, ok := fields[key]
		if !ok {
			continue
		}

		switch field := field.(type) {
		case *string:
			*field = value

		case *net.IP:
			*field = net.ParseIP(value)

		case *time.Time:
			value, err := strconv.ParseFloat(value, 64)
			if err != nil {
				continue
			}
			*field = time.Unix(int64(value), 0)
		}
	}

	return trace, nil
}
