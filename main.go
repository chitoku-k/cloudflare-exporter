package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/chitoku-k/cloudflare-exporter/application/server"
	"github.com/chitoku-k/cloudflare-exporter/infrastructure/cloudflare"
	"github.com/chitoku-k/cloudflare-exporter/infrastructure/config"
	cf "github.com/cloudflare/cloudflare-go"
	"github.com/sirupsen/logrus"
)

var signals = []os.Signal{os.Interrupt}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), signals...)
	defer stop()

	env, err := config.Get()
	if err != nil {
		logrus.Fatalf("Failed to initialize config: %v", err)
	}

	var client *cf.API
	if env.Cloudflare.APIToken == "" {
		client, err = cf.New(env.Cloudflare.APIKey, env.Cloudflare.APIEmail)
	} else {
		client, err = cf.NewWithAPIToken(env.Cloudflare.APIToken)
	}

	if err != nil {
		logrus.Fatalf("Failed to initialize Cloudflare client: %v", err)
	}

	engine := server.NewEngine(
		env.Port,
		env.TLSCert,
		env.TLSKey,
		cloudflare.NewLoadBalancerService(client),
	)
	err = engine.Start(ctx)
	if err != nil {
		logrus.Fatalf("Failed to start web server: %v", err)
	}
}
