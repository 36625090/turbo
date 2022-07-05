package server

import (
	"context"
	"fmt"
	"github.com/36625090/turbo/authorities"
	"github.com/36625090/turbo/config"
	"github.com/36625090/turbo/logical"
	"github.com/36625090/turbo/option"
	"github.com/36625090/turbo/transport"
	"github.com/gin-gonic/gin"
	"github.com/go-various/consul"
	"github.com/hashicorp/go-hclog"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"
)

type Server struct {
	sync.Mutex
	ctx           context.Context
	logger        hclog.InterceptLogger
	opts          *option.Options
	authorization authorities.Authorization
	httpTransport *transport.Transport
	httpServer    *http.Server
	netListener   net.Listener
	connection    *Connection
	backends      map[string]logical.Backend
	service       *consul.Service
	consulClient  consul.Client
	globalConfig  *config.GlobalConfig
	signalChan    chan os.Signal
}

func NewServer(opts *option.Options, cfg *config.GlobalConfig, cl consul.Client, logger hclog.InterceptLogger) *Server {

	gin.SetMode(gin.ReleaseMode)
	en := gin.New()
	if opts.Http.Cors {
		en.Use(Cors())
	}

	return &Server{
		ctx:           context.Background(),
		globalConfig:  cfg,
		opts:          opts,
		consulClient:  cl,
		logger:        logger,
		connection:    &Connection{},
		backends:      map[string]logical.Backend{},
		httpTransport: transport.NewTransport(en, cfg.Transport, logger),
	}
}

//Start the Server
//启动服务
func (m *Server) Start() error {
	m.logger.Info("start starting")
	addr := fmt.Sprintf("%s:%d", m.opts.Http.Address, m.opts.Http.Port)
	m.logger.Info("server listening on ", "address", addr)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	m.netListener = l
	m.httpServer = &http.Server{
		Addr:         addr,
		Handler:      m.httpTransport,
		IdleTimeout:  time.Second * time.Duration(m.opts.Http.IdleTimeout),
		ReadTimeout:  time.Second * time.Duration(m.opts.Http.ReadTimeout),
		WriteTimeout: time.Second * time.Duration(m.opts.Http.WriteTimeout),
		ErrorLog:     m.logger.StandardLogger(&hclog.StandardLoggerOptions{}),
	}
	m.httpServer.SetKeepAlivesEnabled(m.opts.Http.KeepAlive)

	go func() {
		if err := m.httpServer.Serve(m.netListener); err != nil && err != http.ErrServerClosed {
			log.Fatal("start http server: ", err)
			return
		}
	}()
	m.listenHealthyEndpoint()
	if err := m.registerService(m.opts.Profile); err != nil {
		return err
	}

	m.logger.Info("server start completed")
	m.signalChan = make(chan os.Signal)
	signal.Notify(m.signalChan, os.Interrupt)
	<-m.signalChan

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	m.logger.Info("Server shutting")

	m.unRegisterService()

	if err := m.httpServer.Shutdown(ctx); err != nil {
		log.Fatal("server shutdown:", err)
		return err
	}
	m.logger.Info("server shutdown completed")
	m.Cleanup()

	return nil
}

func (m *Server) Stop() {
	if nil == m.signalChan {
		return
	}
	close(m.signalChan)
}

func (m *Server) Cleanup() {
	if nil != m.httpServer {
		m.httpServer.Close()
	}
	if m.netListener != nil {
		m.netListener.Close()
	}
	for _, backend := range m.backends {
		backend.Cleanup(context.Background())
	}
}
