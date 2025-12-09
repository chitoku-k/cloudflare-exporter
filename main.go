package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"

	"github.com/chitoku-k/cloudflare-exporter/application/server"
	"github.com/chitoku-k/cloudflare-exporter/infrastructure/cloudflare"
	"github.com/chitoku-k/cloudflare-exporter/infrastructure/config"
	cf "github.com/cloudflare/cloudflare-go"
	"github.com/spf13/pflag"
)

var (
	signals = []os.Signal{os.Interrupt}
	name    = "cloudflare-exporter"
	version = "v0.0.0-dev"

	flagversion = pflag.BoolP("version", "V", false, "show version")
)

func main() {
	pflag.Parse()
	if *flagversion {
		fmt.Println(name, version)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), signals...)
	defer stop()

	env, err := config.Get()
	if err != nil {
		slog.Error("Failed to initialize config", "err", err)
		os.Exit(1)
	}

	var client *cf.API
	if env.Cloudflare.APIToken == "" {
		client, err = cf.New(env.Cloudflare.APIKey, env.Cloudflare.APIEmail)
	} else {
		client, err = cf.NewWithAPIToken(env.Cloudflare.APIToken)
	}

	if err != nil {
		slog.Error("Failed to initialize Cloudflare client", "err", err)
		os.Exit(1)
	}
	var rc *cf.ResourceContainer
	if env.Cloudflare.AccountID == "" {
		rc = cf.UserIdentifier("")
	} else {
		rc = cf.AccountIdentifier(env.Cloudflare.AccountID)
	}
	engine := server.NewEngine(
		env.Port,
		env.TLSCert,
		env.TLSKey,
		cloudflare.NewLoadBalancerService(client, rc),
		cloudflare.NewProbeService(http.DefaultClient),
	)
	err = engine.Start(ctx)
	if err != nil {
		slog.Error("Failed to start web server", "err", err)
		os.Exit(1)
	}
}
