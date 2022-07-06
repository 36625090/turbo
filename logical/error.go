package logical

import (
	"fmt"
	"github.com/36625090/turbo/logical/codes"
)

type WrapperError struct {
	err   error
	code  codes.ReturnCode
	scope string
}

func NewWrapperError() *WrapperError {
	return &WrapperError{}
}

func (w *WrapperError) WithCode(code codes.ReturnCode) *WrapperError {
	w.code = code
	return w
}
func (w *WrapperError) WithErr(err error) *WrapperError {
	w.err = err
	return w
}
func (w *WrapperError) WithScope(scope string) *WrapperError {
	w.scope = scope
	return w
}
func (w *WrapperError) Code() codes.ReturnCode {
	return w.code
}

func (w *WrapperError) Error() error {
	return w.err
}

func (w *WrapperError) String() string {
	return fmt.Sprintf("WrapperError: code=%d scope=%s err=%s", w.code, w.scope, w.err)
}
