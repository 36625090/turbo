package server

func (m *Server) Initialize(handle func(*TurboContext)) error {
	m.httpTransport.Use(m.requestTracer())

	if m.opts.Newrelic {
		m.httpTransport.Use(m.newrelicTracer())
	}
	m.initBackendAPIServer()

	if m.opts.Ui {
		m.addDocumentSchema()
		m.addDocumentUI()
	}

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
		handle(params)
	}
	return nil
}
