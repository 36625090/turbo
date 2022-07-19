package server

import (
	"context"
	"errors"
	"fmt"
	"github.com/36625090/turbo/authorities"
	"github.com/36625090/turbo/logical"
	"github.com/36625090/turbo/logical/codes"
	"github.com/36625090/turbo/transport"
	"github.com/36625090/turbo/utils"
	"path/filepath"
)

func (m *Server) InitializeBackend(bkName string, factory logical.Factory, cfg *logical.BackendContext) error {
	m.Lock()
	defer m.Unlock()
	if _, ok := m.backends[bkName]; ok {
		return fmt.Errorf("existing backend: %s", bkName)
	}

	backend, err := factory(m.ctx, bkName, cfg)
	if err != nil {
		m.logger.Error("register backend", "name", bkName, "err", err)
		return err
	}

	if err := backend.Initialize(context.Background()); err != nil {
		m.logger.Error("initialize backend", "name", bkName, "err", err)
		return err
	}

	m.backends[bkName] = backend
	m.logger.Info("register backend", "name", bkName)
	return nil
}

func (m *Server) initBackendAPIServer() {

	path := filepath.Join(m.opts.Http.Path, "api")
	if m.opts.Http.Trace {
		m.httpTransport.Use(m.loggerTracker(path))
	}

	m.httpTransport.AddHandle(path, logical.HttpMethodPOST, func(ctx *transport.Context){
		request := ctx.Request()
		m.connection.Inc()
		var err error
		defer func() {
			m.connection.Dec()
			if err != nil {
				m.connection.Error()
			}
		}()

		backend, ok := m.backends[request.Backend()]

		if !ok {
			ctx.WithCode(codes.CodeBackendIssue).WithMessage("invalid backend")
			return
		}

		authorized, err := m.preAuthorization(request.Method, ctx.GetAuthToken())
		if err != nil {
			ctx.WithCode(codes.CodeUnauthorized).WithMessage(err.Error())
			return
		}

		args, err := ctx.DecodeArgs()
		if err != nil {
			ctx.WithCode(codes.CodeFailedDecodeArgs).WithMessage(err.Error())
			return
		}

		args.Authorized = authorized
		resp, werr := backend.HandleRequest(context.Background(), args)
		if werr != nil {
			ctx.WithCode(werr.Code()).WithMessage(werr.Error().Error())
			return
		}

		if resp.Code != 0 {
			ctx.WithCode(codes.ReturnCode(resp.Code)).WithMessage(resp.Message)
			return
		}
		ctx.WithContent(resp.Data)
		ctx.WithPagination(resp.Pagination)
	})

}

func (m *Server) preAuthorization(method string, token string) (*authorities.Authorized, error) {
	if nil == m.authorization {
		return nil, errors.New("authorization unavailable")
	}

	if m.authorization.Settings().DefaultPolicy == authorities.AuthorizationPolicyAllow {
		return nil, nil
	}

	if utils.Contains(m.authorization.Settings().AnonMethods, method) {
		return nil, nil
	}

	if token == "" {
		return nil, errors.New("invalid token")
	}

	return m.authorization.Authentication(context.TODO(), token)
}
