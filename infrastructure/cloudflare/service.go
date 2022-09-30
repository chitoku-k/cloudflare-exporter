package cloudflare

import (
	"context"
	"fmt"

	"github.com/chitoku-k/cloudflare-exporter/service"
	cf "github.com/cloudflare/cloudflare-go"
)

type loadBalancerService struct {
	Client *cf.API
}

func NewLoadBalancerService(client *cf.API) service.LoadBalancer {
	return &loadBalancerService{
		Client: client,
	}
}

func (s *loadBalancerService) Collect(ctx context.Context, poolName string) ([]service.Pool, error) {
	pools, err := s.Client.ListLoadBalancerPools(ctx, cf.UserIdentifier(""), cf.ListLoadBalancerPoolParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to list pools: %w", err)
	}

	var result []service.Pool
	for _, p := range pools {
		if p.Name != poolName {
			continue
		}

		health, err := s.Client.GetLoadBalancerPoolHealth(ctx, cf.UserIdentifier(""), p.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get pool health: %w", err)
		}

		var healths []service.PopHealth
		for region, pop := range health.PopHealth {
			var origins []service.Origin
			for _, originHealths := range pop.Origins {
				for address, h := range originHealths {
					origins = append(origins, service.Origin{
						Address:      address,
						Healthy:      h.Healthy,
						RTT:          h.RTT.Duration,
						ResponseCode: h.ResponseCode,
					})
				}
			}

			healths = append(healths, service.PopHealth{
				Region:  region,
				Healthy: pop.Healthy,
				Origins: origins,
			})
		}

		result = append(result, service.Pool{
			Name:       p.Name,
			PopHealths: healths,
		})
	}

	return result, nil
}
