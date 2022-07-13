package logical

import (
	"reflect"
	"testing"
)

type Member struct {
	Name string
}

type Members []Member
type MemberMap map[string]Member
func TestFields(t *testing.T) {
	fields := getFields(reflect.TypeOf(Members{}))
	t.Log(fields[0])
	fields = getFields(reflect.TypeOf(MemberMap{}))
	t.Log(fields[0])
}
