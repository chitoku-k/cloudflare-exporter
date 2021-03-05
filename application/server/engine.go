package server

import (
	"context"
	"fmt"
	"net"
	"net/http"

	"github.com/chitoku-k/cloudflare-exporter/service"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

type engine struct {
	Port         string
	LoadBalancer service.LoadBalancer
}

type Engine interface {
	Start(ctx context.Context) error
}

func NewEngine(
	port string,
	loadBalancer service.LoadBalancer,
) Engine {
	return &engine{
		Port:         port,
		LoadBalancer: loadBalancer,
	}
}

func (e *engine) Start(ctx context.Context) error {
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(gin.LoggerWithConfig(gin.LoggerConfig{
		Formatter: e.Formatter(),
		SkipPaths: []string{"/healthz"},
	}))

	router.Any("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	router.GET("/metrics", func(c *gin.Context) {
		health := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "cloudflare",
			Name:      "origin_health",
			Help:      "Result of health check",
		}, []string{"pool_name", "health_region", "origin_address", "code"})

		rtt := prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "cloudflare",
			Name:      "origin_rtt_seconds",
			Help:      "RTT to the pool origin",
		}, []string{"pool_name", "health_region", "origin_address", "code"})

		target, ok := c.GetQuery("target")
		if !ok {
			c.Status(http.StatusBadRequest)
			return
		}

		pools, err := e.LoadBalancer.Collect(c, target)
		if err != nil {
			logrus.Errorf("Error in Cloudflare: %v", err)
			c.Status(http.StatusInternalServerError)
			return
		}

		for _, p := range pools {
			for _, h := range p.PopHealths {
				for _, o := range h.Origins {
					var value float64
					if o.Healthy {
						value = 1
					}

					labels := prometheus.Labels{
						"pool_name":      p.Name,
						"health_region":  h.Region,
						"origin_address": o.Address,
						"code":           fmt.Sprint(o.ResponseCode),
					}
					health.With(labels).Set(value)
					rtt.With(labels).Set(o.RTT.Seconds())
				}
			}
		}

		registry := prometheus.NewRegistry()
		registry.MustRegister(health, rtt)

		handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
		handler.ServeHTTP(c.Writer, c.Request)
	})

	server := http.Server{
		Addr:    net.JoinHostPort("", e.Port),
		Handler: router,
	}

	var eg errgroup.Group
	eg.Go(func() error {
		<-ctx.Done()
		return server.Shutdown(context.Background())
	})

	err := server.ListenAndServe()
	if err == http.ErrServerClosed {
		return eg.Wait()
	}

	return err
}

func (e *engine) Formatter() gin.LogFormatter {
	return func(param gin.LogFormatterParams) string {
		remoteHost, _, err := net.SplitHostPort(param.Request.RemoteAddr)
		if remoteHost == "" || err != nil {
			remoteHost = "-"
		}

		bodySize := fmt.Sprintf("%v", param.BodySize)
		if param.BodySize == 0 {
			bodySize = "-"
		}

		referer := param.Request.Header.Get("Referer")
		if referer == "" {
			referer = "-"
		}

		userAgent := param.Request.Header.Get("User-Agent")
		if userAgent == "" {
			userAgent = "-"
		}

		forwardedFor := param.Request.Header.Get("X-Forwarded-For")
		if forwardedFor == "" {
			forwardedFor = "-"
		}

		return fmt.Sprintf(`%s %s %s [%s] "%s %s %s" %v %s "%s" "%s" "%s"%s`,
			remoteHost,
			"-",
			"-",
			param.TimeStamp.Format("02/Jan/2006:15:04:05 -0700"),
			param.Request.Method,
			param.Request.RequestURI,
			param.Request.Proto,
			param.StatusCode,
			bodySize,
			referer,
			userAgent,
			forwardedFor,
			"\n",
		)
	}
}
