package server

func (m *Server) Initialize(handle func(*TurboContext)) error {
	params := &TurboContext{
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
