package logical

import (
	"reflect"
	"testing"
	"time"
)

type Member struct {
	MemberName string
}

type Members []time.Time
type Members2 []*Member
type MemberMap map[string]Member
type MemberMap2 map[string]*Member

func TestFields(t *testing.T) {
	fields := getFields(reflect.TypeOf(Members{}))
	t.Log(fields[0])
	fields = getFields(reflect.TypeOf(Members2{}))
	t.Log(fields[0])
	//fields = getFields(reflect.TypeOf(MemberMap{}))
	//t.Log(fields[0])
}
