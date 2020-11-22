package service

import (
	"context"
)

type HealthCheck interface {
	Do(ctx context.Context, nodeName string) (bool, error)
}
