package controllers

import (
	echoPrometheus "github.com/globocom/echo-prometheus"
	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

// ConfigManagerController implements ServerInterface
type MetricsServer struct {
	Server *echo.Echo
	Config *viper.Viper
}

func NewMetricsServer(cfg *viper.Viper) *MetricsServer {
	return &MetricsServer{
		Server: echo.New(),
		Config: cfg,
	}
}

// Routes sets up middlewares and registers handlers for each route
func (s *MetricsServer) Routes() {
	s.Server.Use(echoPrometheus.MetricsMiddleware())
	s.Server.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
}

// Start starts an http server with addr
func (s *MetricsServer) Start(addr string) {
	s.Server.Start(addr)
}
