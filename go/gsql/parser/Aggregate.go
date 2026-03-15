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

// Aggregate.go provides parsing support for aggregate functions in L8QL SELECT clauses.
// Supported functions: count(*), count(field), sum(field), avg(field), min(field), max(field).
package parser

import (
	"github.com/saichler/l8types/go/types/l8api"
	"strings"
)

// aggregateFunctions lists the supported aggregate function names.
var aggregateFunctions = []string{"count", "sum", "avg", "min", "max"}

// parseAggregateFunction detects if a SELECT column is an aggregate function call.
// Returns the parsed L8AggregateFunction and true if it is, or nil and false otherwise.
// Examples: "count(*)" -> {function:"count", field:"*", alias:"count"}
//
//	"sum(salary)" -> {function:"sum", field:"salary", alias:"sumSalary"}
func parseAggregateFunction(col string) (*l8api.L8AggregateFunction, bool) {
	col = strings.TrimSpace(col)
	lower := strings.ToLower(col)

	for _, fn := range aggregateFunctions {
		prefix := fn + "("
		if strings.HasPrefix(lower, prefix) && strings.HasSuffix(lower, ")") {
			// Extract field name between parentheses
			field := strings.TrimSpace(col[len(prefix) : len(col)-1])
			if field == "" {
				continue
			}

			alias := buildAlias(fn, field)
			return &l8api.L8AggregateFunction{
				Function: fn,
				Field:    field,
				Alias:    alias,
			}, true
		}
	}
	return nil, false
}

// buildAlias generates a display alias for an aggregate function.
// count(*) -> "count", sum(salary) -> "sumSalary", avg(amount) -> "avgAmount"
func buildAlias(fn, field string) string {
	if field == "*" {
		return fn
	}
	// Capitalize first letter of field
	if len(field) > 0 {
		return fn + strings.ToUpper(field[:1]) + field[1:]
	}
	return fn + field
}

// isAggregateQuery checks if any property in the SELECT clause is an aggregate function.
func isAggregateQuery(props []string) bool {
	for _, prop := range props {
		_, ok := parseAggregateFunction(prop)
		if ok {
			return true
		}
	}
	return false
}
