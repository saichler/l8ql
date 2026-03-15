/*
© 2025 Sharon Aicler (saichler@gmail.com)

Layer 8 Ecosystem is licensed under the Apache License, Version 2.0.
You may obtain a copy of the License at:

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package tests

import (
	"testing"

	. "github.com/saichler/l8ql/go/gsql/parser"
	. "github.com/saichler/l8test/go/infra/t_resources"
)

// TestParseCountStar tests parsing count(*) in SELECT clause.
func TestParseCountStar(t *testing.T) {
	q, e := NewQuery("select count(*) from TestProto", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	if len(q.Query().Aggregates) != 1 {
		Log.Fail(t, "Expected 1 aggregate function")
		return
	}
	agg := q.Query().Aggregates[0]
	if agg.Function != "count" || agg.Field != "*" || agg.Alias != "count" {
		Log.Fail(t, "Unexpected aggregate:", agg.Function, agg.Field, agg.Alias)
	}
}

// TestParseSumField tests parsing sum(field) in SELECT clause.
func TestParseSumField(t *testing.T) {
	q, e := NewQuery("select sum(myInt32) from TestProto", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	if len(q.Query().Aggregates) != 1 {
		Log.Fail(t, "Expected 1 aggregate function")
		return
	}
	agg := q.Query().Aggregates[0]
	if agg.Function != "sum" || agg.Field != "myInt32" || agg.Alias != "sumMyInt32" {
		Log.Fail(t, "Unexpected aggregate:", agg.Function, agg.Field, agg.Alias)
	}
}

// TestParseMultipleAggregates tests parsing multiple aggregate functions.
func TestParseMultipleAggregates(t *testing.T) {
	q, e := NewQuery("select myString,count(*),avg(myInt32) from TestProto group-by myString", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	if len(q.Query().Aggregates) != 2 {
		Log.Fail(t, "Expected 2 aggregate functions, got", len(q.Query().Aggregates))
		return
	}
	// myString should remain in Properties (non-aggregate)
	if len(q.Query().Properties) != 1 || q.Query().Properties[0] != "myString" {
		Log.Fail(t, "Expected myString in Properties")
		return
	}
	// group-by should be parsed
	if len(q.Query().GroupBy) != 1 || q.Query().GroupBy[0] != "myString" {
		Log.Fail(t, "Expected group-by myString")
	}
}

// TestParseGroupBy tests parsing GROUP BY clause.
func TestParseGroupBy(t *testing.T) {
	q, e := NewQuery("select myString,count(*) from TestProto group-by myString", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	if len(q.Query().GroupBy) != 1 {
		Log.Fail(t, "Expected 1 group-by field")
		return
	}
	if q.Query().GroupBy[0] != "myString" {
		Log.Fail(t, "Expected group-by field myString, got", q.Query().GroupBy[0])
	}
}

// TestParseNonAggregateUnchanged verifies non-aggregate queries still work.
func TestParseNonAggregateUnchanged(t *testing.T) {
	q, e := NewQuery("select myString,myInt32 from TestProto where myBool=true", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	if len(q.Query().Aggregates) != 0 {
		Log.Fail(t, "Expected 0 aggregates for non-aggregate query")
		return
	}
	if len(q.Query().Properties) != 2 {
		Log.Fail(t, "Expected 2 properties")
		return
	}
	if len(q.Query().GroupBy) != 0 {
		Log.Fail(t, "Expected 0 group-by fields")
	}
}

// TestAggregateExecution tests executing aggregate functions against data.
func TestAggregateExecution(t *testing.T) {
	q, _, e := createQuery("select myString,count(*),sum(myInt32),avg(myInt32) from TestProto group-by myString")
	if e != nil {
		Log.Fail(t, e)
		return
	}
	if !q.IsAggregate() {
		Log.Fail(t, "Expected aggregate query")
		return
	}

	// Create test data with different myString values
	items := make([]interface{}, 0)
	for i := 0; i < 6; i++ {
		node := CreateTestModelInstance(i)
		if i < 3 {
			node.MyString = "GroupA"
		} else {
			node.MyString = "GroupB"
		}
		node.MyInt32 = int32(10 * (i + 1))
		items = append(items, node)
	}

	results := q.Aggregate(items)
	if len(results) != 2 {
		Log.Fail(t, "Expected 2 groups, got", len(results))
		return
	}

	// Verify counts
	for _, r := range results {
		count := r["count"].(int64)
		if count != 3 {
			Log.Fail(t, "Expected count 3, got", count)
		}
	}
}

// TestAggregateWithFilter tests aggregate with WHERE filtering.
func TestAggregateWithFilter(t *testing.T) {
	q, _, e := createQuery("select count(*) from TestProto where myInt32>20")
	if e != nil {
		Log.Fail(t, e)
		return
	}

	items := make([]interface{}, 0)
	for i := 0; i < 5; i++ {
		node := CreateTestModelInstance(i)
		node.MyInt32 = int32(10 * (i + 1)) // 10, 20, 30, 40, 50
		items = append(items, node)
	}

	// Filter first, then aggregate
	filtered := q.Filter(items, false)
	results := q.Aggregate(filtered)
	if len(results) != 1 {
		Log.Fail(t, "Expected 1 group (no group-by)")
		return
	}
	count := results[0]["count"].(int64)
	if count != 3 {
		Log.Fail(t, "Expected count 3 (values > 20), got", count)
	}
}

// TestParseAllAggregateFunctions verifies all 5 aggregate functions parse correctly.
func TestParseAllAggregateFunctions(t *testing.T) {
	q, e := NewQuery("select count(*),sum(myInt32),avg(myInt32),min(myInt32),max(myInt32) from TestProto", Log)
	if e != nil {
		Log.Fail(t, e)
		return
	}
	if len(q.Query().Aggregates) != 5 {
		Log.Fail(t, "Expected 5 aggregate functions, got", len(q.Query().Aggregates))
		return
	}
	expected := []string{"count", "sum", "avg", "min", "max"}
	for i, agg := range q.Query().Aggregates {
		if agg.Function != expected[i] {
			Log.Fail(t, "Expected function", expected[i], "got", agg.Function)
		}
	}
}
