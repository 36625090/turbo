package transport

import (
	"errors"
)

var errInvalidSign = errors.New("invalid signature")
var errInvalidHeaderSignKey = errors.New("headers[X-Client-ID] cannot be empty")
var errInvalidSignKey = errors.New("invalid signature key id")
var errInvalidDefaultSignKey = errors.New("invalid signature global default key")

type Signer interface {
	Sign(keyId string, resp Codec) (string, error)
	Verify(keyId, sign string, req Codec) error
}
