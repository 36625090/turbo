package authorities

import (
	"encoding/json"
	"strconv"
)

type ID string

func (id ID) Int() int {
	i, _ := strconv.Atoi(string(id))
	return i
}
func (id ID) Int64() int64 {
	i, _ := strconv.ParseInt(string(id), 10, 64)
	return i
}

func (id ID) String() string {
	return string(id)
}

//Authorized
//验证信息
type Authorized struct {
	ID        ID        `json:"id" name:"账户ID"`
	Account   string    `json:"account" name:"账户名称"`
	Principal Principal `json:"principal" name:"账户凭证(用户信息)"`
}

type Principal map[string]interface{}

func (m Principal) Get(key string) interface{} {
	if nil == m {
		return nil
	}
	return m[key]
}

func NewAuthorized(id, account string, principal Principal) Authorized {
	return Authorized{ID: ID(id), Account: account, Principal: principal}
}

func (a Authorized) Encode() ([]byte, error) {
	return json.Marshal(a)
}

func (a Authorized) GetPrincipal() Principal {
	return a.Principal
}

func (a Authorized) SetPrincipal(in Principal) {
	a.Principal = in
}
