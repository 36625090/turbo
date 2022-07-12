package logical

import (
	"encoding/json"
	"github.com/36625090/turbo/authorities"
	"github.com/36625090/turbo/logical/codes"
	"github.com/36625090/turbo/utils"
	"github.com/go-playground/validator/v10"
)

var validate *validator.Validate

func init() {
	validate = validator.New()
}

type Args struct {
	Backend    string                  `json:"backend" validate:"required"`
	Endpoint   string                  `json:"endpoint" validate:"required"`
	Operation  string                  `json:"operation" validate:"required"`
	Data       interface{}             `json:"data" validate:"required"`
	Authorized *authorities.Authorized `json:"authorized"`
	Token      string                  `json:"token"`
	Headers    map[string][]string     `json:"headers"`
	Connection *Connection             `json:"connection" validate:"required"`
}

func (r *Args) GetTraceID() string {
	traces, ok := r.Headers[string(HeaderTraceIDKey)]
	if ok && len(traces) > 0 {
		return traces[0]
	}
	return ""
}

func (r *Args) SetTraceID(id string) *Args {
	r.Headers[string(HeaderTraceIDKey)] = []string{id}
	return r
}

func (r *Args) ShouldBindJSON(out interface{}) *WrapperError {
	if err := json.Unmarshal([]byte(r.Data.(string)), out); err != nil {
		return NewWrapperError().WithErr(err).
			WithCode(codes.CodeBindRequestData)
	}
	if err := validate.Struct(out); err != nil {
		return NewWrapperError().WithErr(err).
			WithCode(codes.CodeBindRequestData)
	}
	return nil
}

func (r *Args) String() string {
	return utils.JSONDump(r)
}
