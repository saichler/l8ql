package tests

import (
	"github.com/saichler/l8pollaris/go/types/l8tpollaris"
	"github.com/saichler/l8ql/go/gsql/interpreter"
	. "github.com/saichler/l8test/go/infra/t_resources"
	"github.com/saichler/l8types/go/ifs"
	"testing"
)

func TestSpecialCase(t *testing.T) {
	r, _ := CreateResources(25000, 2, ifs.Trace_Level)
	r.Introspector().Decorators().AddPrimaryKeyDecorator(&l8tpollaris.L8PTarget{}, "TargetId")
	gsql := "select * from L8PTarget where InventoryType=1 and (State=0 or State=1)"
	_, e := interpreter.NewQuery(gsql, r)
	if e != nil {
		Log.Fail(t, e)
		return
	}
}

func TestSortBy(t *testing.T) {
	r, _ := CreateResources(25000, 2, ifs.Trace_Level)
	r.Introspector().Decorators().AddPrimaryKeyDecorator(&l8tpollaris.L8PTarget{}, "TargetId")
	gsql := "select * from L8PTarget where InventoryType=1 and (State=0 or State=1) sort-by hosts.hostid"
	_, e := interpreter.NewQuery(gsql, r)
	if e != nil {
		Log.Fail(t, e)
		return
	}
}
