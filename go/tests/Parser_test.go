package tests

import (
	"fmt"
	"strconv"
	"testing"

	. "github.com/saichler/l8ql/go/gsql/parser"
	. "github.com/saichler/l8test/go/infra/t_resources"
)

func TestQuery01(t *testing.T) {
	q, e := NewQuery("Select column1,column2 fRom table1 wHere 1=2", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	testTables(q, "table1", t)
	testColumns(q, []string{"column1", "column2"}, t)
	testExpression(q, "(1=2)", t)
}

func TestQuery02(t *testing.T) {
	q, e := NewQuery("Select column1 fRom table1 wHere 1=2", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	testTables(q, "table1", t)
	testColumns(q, []string{"column1"}, t)
	testExpression(q, "(1=2)", t)
}

func TestQuery03(t *testing.T) {
	q, e := NewQuery("Select column1,column2 fRom table1 wHere 1=2 AND 3=4", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	testTables(q, "table1", t)
	testColumns(q, []string{"column1", "column2"}, t)
	testExpression(q, "(1=2 and 3=4)", t)
}

func TestQuery04(t *testing.T) {
	q, e := NewQuery("Select column1,column2 fRom table1 wHere 1=2 AND 3  =  4", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	testTables(q, "table1", t)
	testColumns(q, []string{"column1", "column2"}, t)
	testExpression(q, "(1=2 and 3=4)", t)
}

func TestQuery05(t *testing.T) {
	q, e := NewQuery("Select column1,column2 fRom table1 wHere 1=2 AND 3  =  4 Or 5!=6", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	testTables(q, "table1", t)
	testColumns(q, []string{"column1", "column2"}, t)
	testExpression(q, "(1=2 and 3=4 or 5!=6)", t)
}

func TestQuery06(t *testing.T) {
	q, e := NewQuery("Select column1,column2 fRom table1 wHere 1=2 AND (3  =  4 Or 5!=6)", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	testTables(q, "table1", t)
	testColumns(q, []string{"column1", "column2"}, t)
	testExpression(q, "(1=2) and ((3=4 or 5!=6))", t)
}

func TestQuery07(t *testing.T) {
	q, e := NewQuery("Select column1,column2 fRom table1 wHere (1=2 or 3  =  4) And (5!=6 or 8<9) or 10<=12", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	testTables(q, "table1", t)
	testColumns(q, []string{"column1", "column2"}, t)
	testExpression(q, "((1=2 or 3=4)) and ((5!=6 or 8<9)) or (10<=12)", t)
}

func TestQuery08(t *testing.T) {
	_, e := NewQuery("Select column1,column2 fRom table1 wHere (1=2 or 3  =  4) And (5!=6 or 8<9 or 10<=12", Log)
	if e == nil {
		Log.Fail(t, "Expected Fail")
		return
	}
}

func TestQuery09(t *testing.T) {
	_, e := NewQuery("Select column1,column2 fRom table1 wHere (1=2 or 3  =  4) And 5!=6 or 8<9) or 10<=12", Log)
	if e == nil {
		Log.Fail(t, "Expected a failure")
		return
	}
}

func TestQuery10(t *testing.T) {
	_, e := NewQuery("Select column1,column2 fRom table1 wHere (1=2 or 3  =  4) Anf (5!=6 or 8<9) or 10<=12", Log)
	if e == nil {
		Log.Fail(t, "Expected fail")
		return
	}
}

func TestQuery11(t *testing.T) {
	_, e := NewQuery("Select column1,column2 fRom table1 wHere (1=2 or 3  =  4) And (5^6 or 8<9) or 10<=12", Log)
	if e == nil {
		Log.Fail(t, "Expected fail")
		return
	}
}

func TestQuery12(t *testing.T) {
	q, e := NewQuery("Select column1,column2 fRom table1 wHere (1=2 or 3  =  4) And (5!=6 or 8<9) or 10<=12 sort-by col1 page 7 limit 50 match-case descending", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	if !q.Query().MatchCase {
		Log.Fail(t, "Expected match-case to match")
		return
	}
	if !q.Query().Descending {
		Log.Fail(t, "Expected Descending to be true")
		return
	}
	if q.Query().SortBy != "col1" {
		Log.Fail(t, "Expected sort-by to be col1")
		return
	}
	if q.Query().Page != 7 {
		Log.Fail(t, "Expected page to be 7")
		return
	}
	if q.Query().Limit != 50 {
		Log.Fail(t, "Expected kimit to be 50")
		return
	}

	testTables(q, "table1", t)
	testColumns(q, []string{"column1", "column2"}, t)
	testExpression(q, "((1=2 or 3=4)) and ((5!=6 or 8<9)) or (10<=12)", t)
}

func TestVisualize(t *testing.T) {
	q, e := NewQuery("Select column1,column2 fRom table1 wHere (1=2 or 3  =  4) And (5!=6 or 8<9) or 10<=12", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	fmt.Println(VisualizeExpression(q.Query().Criteria, 0))
}

func TestQuery(t *testing.T) {
	q, e := NewQuery("Select column1,column2 fRom table1 wHere 1=2 or ((3!=4 and 5<6) and 7>8) or ((9=10) and 11=12) ", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	testTables(q, "table1", t)
	testColumns(q, []string{"column1", "column2"}, t)
	testExpression(q, "(1=2) or (((3!=4 and 5<6)) and (7>8)) or (((9=10)) and (11=12))", t)
}

func testTables(q *PQuery, expected string, t *testing.T) {
	if q.Query().RootType == "" {
		Log.Fail(t, "Expected ", expected)
		return
	}
}

func testColumns(q *PQuery, expected []string, t *testing.T) {
	if len(q.Query().Properties) != len(expected) {
		Log.Fail(t, "Expected "+strconv.Itoa(len(expected)), " columns but got ", strconv.Itoa(len(q.Query().Properties)))
		return
	}
	for _, et := range expected {
		found := false
		for _, qc := range q.Query().Properties {
			if qc == et {
				found = true
				break
			}
		}
		if !found {
			Log.Fail(t, "Expected column ", et, " but did not find it")
			return
		}
	}
}

func testExpression(q *PQuery, expected string, t *testing.T) {
	if StringExpression(q.Query().Criteria) != expected {
		Log.Fail(t, "Expected: ", expected)
		Log.Fail(t, "But got : ", StringExpression(q.Query().Criteria))
	}
}

func TestParseMapReduce(t *testing.T) {
	q, e := NewQuery("select * from l8file where path=\"*\" mapreduce true", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	v := q.Query().Criteria.Condition.Comparator.Right
	if v != "*" {
		Log.Fail(t, v)
		return
	}
}
