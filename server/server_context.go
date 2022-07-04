package server

import (
	"context"
	"fmt"
	"github.com/36625090/turbo/config"
	"github.com/36625090/turbo/logical"
	"github.com/36625090/turbo/transport"
	"github.com/gin-gonic/gin"
	"github.com/go-various/consul"
	"github.com/google/uuid"
	"github.com/hashicorp/go-hclog"
	"time"
)


type TurboContext struct {
	Context  context.Context
	Backends map[string]logical.Backend
	Config   *config.GlobalConfig
	Consul   consul.Client
	Logger   hclog.InterceptLogger
	Transport *transport.Transport
}

func (m *Server) requestTracer() gin.HandlerFunc {

	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			return
		}
		// 设置 trace-id 变量
		c.Request.Header.Set(string(logical.HeaderApplicationKey), m.opts.App)
		c.Request.Header.Set(string(logical.HeaderTraceIDKey), uuid.New().String())

		c.Next()
	}
}

func (m *Server) loggerTracker(path string) gin.HandlerFunc {
	return func(c *gin.Context) {

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		t := time.Now()
		defer func() {
			// 请求后
			latency := time.Since(t)
			if c.Request.RequestURI != path {
				return
			}
			msg := fmt.Sprintf("client=%v client-id=%s trace-id=%s application=%s uri=%v code=%d latency=%v",
				c.Request.RemoteAddr,
				c.Request.Header.Get(logical.HeaderClientIDKey.String()),
				c.Request.Header.Get(logical.HeaderTraceIDKey.String()),
				c.Request.Header.Get(logical.HeaderApplicationKey.String()),
				c.Request.RequestURI, c.Writer.Status(), latency)
			m.logger.Info(msg)
		}()
		defer c.Next()

	}
}
