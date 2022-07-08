package transport

import (
	"fmt"
	"github.com/36625090/turbo/logical"
	"github.com/36625090/turbo/logical/codes"
	"github.com/36625090/turbo/utils"
	"github.com/gin-gonic/gin"
	"github.com/hashicorp/go-hclog"
	"runtime/debug"
)

//Transport 继承了gin实现的服务接口
type Transport struct {
	*gin.Engine
	logger hclog.Logger
	signer Signer
}

type Handle func(c *Context)

func NewTransport(en *gin.Engine, settings *Settings, logger hclog.InterceptLogger) *Transport {
	transport := &Transport{
		Engine: en,
		logger: logger,
		signer: NewMD5Signer(settings, logger),
	}
	return transport
}

//AddHandle 添加路径handlerFunc
//path 绝对路径
func (m *Transport) AddHandle(absolutePath string, method logical.HttpMethod, handle Handle) {
	m.logger.Info("initialize handle", "path", absolutePath, "method", method)
	m.Engine.Handle(string(method), absolutePath, func(gCtx *gin.Context) {
		ctx := NewContext(gCtx)
		ctx.ctx = gCtx
		ctx.response.TraceID = ctx.GetTraceID()

		defer func() {
			if r := recover(); r != nil {
				m.logger.Error("received panic", "err", r, "stack", string(debug.Stack()))
				ctx.WithCode(codes.CodeHandleRequest).
					WithMessage(fmt.Sprintf("%v", r))
				ctx.write()
			}
		}()

		if err := ctx.ShouldBind(); err != nil {
			m.logger.Error("should not bind JSON", "path", ctx.RawRequest().RequestURI, "err", err)

			ctx.WithCode(codes.CodeBindRequestData).
				WithMessage(err.Error()).write()
			return
		}

		if err := m.signer.Verify(ctx.GetClientID(), ctx.request.Sign, ctx.request); err != nil {
			m.logger.Error("verify request sign error",
				"path", ctx.RawRequest().RequestURI,
				"client", ctx.GetClientID(),
				"sign", ctx.request.Sign,
				"original", utils.JSONDump(ctx.request),
				"err", err)
			ctx.WithCode(codes.CodeInvalidSignature).
				WithMessage("verify request sign error, " + err.Error() + " : " + ctx.Request().Sign).write()
			return
		}

		handle(ctx)
		sign, err := m.signer.Sign(GlobalSignKey, ctx.response)
		if err != nil {
			ctx.WithCode(codes.CodeInvalidSignature).WithMessage(err.Error()).write()
			return
		}
		ctx.WithSign(sign)
		ctx.write()

	})
}

func (m *Transport) Router() gin.IRouter {
	return m.Engine
}

func (m *Transport) SetSigner(signer Signer) {
	m.signer = signer
}
