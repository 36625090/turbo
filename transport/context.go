package transport

import (
	"github.com/36625090/turbo/logical"
	"github.com/36625090/turbo/logical/codes"
	"github.com/gin-gonic/gin"
	"github.com/pkg/errors"
	"net/http"
	"strings"
	"time"
)

type Context struct {
	ctx      *gin.Context
	request  *Request
	response *Response
}

func NewContext(ctx *gin.Context) *Context {
	context := &Context{
		ctx:     ctx,
		request: new(Request),
		response: &Response{
			Message:   "",
			Content:   nil,
			Timestamp: time.Now().UnixMilli(),
			Sign:      "00000000000000000000000000000000",
		},
	}
	return context
}

func (c *Context) GetTraceID() string {
	return c.ctx.GetHeader(string(logical.HeaderTraceIDKey))
}

func (c *Context) GetAuthToken() string {
	return c.ctx.GetHeader(string(logical.HeaderAuthorizationKey))
}

func (c *Context) GetClientID() string {
	return c.ctx.GetHeader(string(logical.HeaderClientIDKey))
}

func (c *Context) ShouldBind() error {
	return c.ctx.ShouldBind(c.request)
}

//Request 获取客户端请求数据
func (c *Context) Request() *Request {
	return c.request
}

func (c *Context) RawRequest() *http.Request {
	return c.ctx.Request
}

func (c *Context) RawWriter() gin.ResponseWriter {
	return c.ctx.Writer
}

func (c *Context) RawParams() gin.Params {
	return c.ctx.Params
}

func (c *Context) RawContext() *gin.Context {
	return c.ctx
}

//DecodeArgs 解码逻辑请求数据结构
func (c *Context) DecodeArgs() (*logical.Args, error) {

	methods := strings.Split(c.request.Method, ".")
	if len(methods) != 3 {
		return nil, errors.New("method not supported: " + c.request.Method)
	}

	//逻辑请求
	args := &logical.Args{
		Backend:    methods[0],
		Endpoint:   methods[1],
		Operation:  methods[2],
		Data:       c.request.Data,
		Headers:    c.ctx.Request.Header,
		Connection: &logical.Connection{RemoteAddr: c.ctx.Request.RemoteAddr, UserAgent: c.ctx.Request.UserAgent()},
	}

	args.SetTraceID(c.GetTraceID())

	return args, nil

}

func (c *Context) WithCode(code codes.ReturnCode) *Context {
	c.response.Code = code.Int()
	return c
}

func (c *Context) WithMessage(message string) *Context {
	c.response.Message = message
	return c
}

func (c *Context) WithContent(content interface{}) *Context {
	c.response.Content = content
	return c
}
func (c *Context) WithPagination(pagination interface{}) *Context {
	c.response.Pagination = pagination
	return c
}

func (c *Context) WithSign(sign string) *Context {
	c.response.Sign = sign
	return c
}

func (c *Context) write() {
	c.ctx.JSON(200, c.response)
}
