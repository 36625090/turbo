package server

import (
	"github.com/36625090/turbo/authorities"
	"github.com/36625090/turbo/transport"
)

func (m *Server) Initialize(handle func(*TurboContext) error) error {
	params := &TurboContext{
		Options:   m.opts,
		Context:   m.ctx,
		Backends:  m.backends,
		Config:    m.globalConfig,
		Consul:    m.consulClient,
		Logger:    m.logger,
		Transport: m.httpTransport,
	}
	if handle != nil {
		if err := handle(params); err != nil {
			return err
		}
	}

	m.httpTransport.Use(m.requestTracer())

	if m.opts.Newrelic {
		m.httpTransport.Use(m.newrelicTracer())
	}
	m.initBackendAPIServer()

	if m.opts.Ui {
		m.addDocumentSchema()
		m.addDocumentUI()
	}
	return nil
}

//InitializeAuthorization 注册验证代理接口，如不需要课不注册
func (m *Server) InitializeAuthorization(authorization authorities.Authorization) error {
	m.authorization = authorization
	return nil
}

//InitializeSigner 注册加签验签
func (m *Server) InitializeSigner(signer transport.Signer) {
	m.httpTransport.SetSigner(signer)
}
