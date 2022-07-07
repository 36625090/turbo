package transport

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/36625090/turbo/utils"
	"github.com/hashicorp/go-hclog"
)

type signer struct {
}

func (m *signer) build(req Codec) *bytes.Buffer {
	params := req.Map()
	var buf bytes.Buffer
	for _, v := range req.Keys() {
		if v == "sign" {
			continue
		}
		if utils.IsNil(params[v]) {
			continue
		}
		buf.WriteString(v)
		buf.WriteString(fmt.Sprintf("%v", params[v]))
	}
	return &buf
}

type md5Signer struct {
	signer
	settings *Settings
	logger   hclog.Logger
}

func NewMD5Signer(settings *Settings, logger hclog.Logger) Signer {
	return &md5Signer{
		settings: settings,
		logger:   logger,
	}
}

//Sign 签名方法
//keyId keys id
func (m *md5Signer) Sign(keyId string, resp Codec) (string, error) {
	if m.settings.DefaultPolicy == SignPolicyAllow {
		if m.logger.IsTrace() {
			m.logger.Trace("md5 signer default policy is allowed")
		}
		return "00000000000000000000000000000000", nil
	}

	key := m.settings.SignKeys[keyId]
	if "" == key {
		if m.logger.IsTrace() {
			m.logger.Trace("md5 signer locate key", "id", keyId, "key", key)
		}
		return "", errInvalidDefaultSignKey
	}

	buf := m.build(resp)
	buf.WriteString(key)
	sign := m.md5(buf.Bytes())
	if m.logger.IsTrace() {
		m.logger.Trace("md5 signature", "original", buf.String(), "sign", sign)
	}
	return sign, nil
}

//Verify 签名校验方法
func (m *md5Signer) Verify(keyId, sign string, req Codec) error {
	if m.settings.DefaultPolicy == SignPolicyAllow {
		return nil
	}

	if "" == keyId {
		return errInvalidHeaderSignKey
	}

	key := m.settings.SignKeys[keyId]
	if "" == key {
		return errors.New(errInvalidSignKey.Error() + ": " + keyId)
	}

	buf := m.build(req)
	buf.WriteString(key)
	newSign := m.md5(buf.Bytes())
	if m.logger.IsTrace() {
		m.logger.Trace("signature verify", "original", sign, "calculated", newSign)
	}
	if newSign != sign {
		return errInvalidSign
	}
	return nil
}

func (m *md5Signer) md5(data []byte) string {
	md := md5.New()
	md.Write(data)
	return hex.EncodeToString(md.Sum(nil))
}
