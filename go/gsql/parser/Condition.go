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
package parser

import (
	"bytes"
	"errors"
	"strings"

	"github.com/saichler/l8types/go/types/l8api"
)

// ConditionOperation represents the logical operators used to combine conditions in a WHERE clause.
type ConditionOperation string

// Logical operators for combining conditions and expression size limit.
const (
	And                 ConditionOperation = " and " // Logical AND operator
	Or                  ConditionOperation = " or "  // Logical OR operator
	MAX_EXPRESSION_SIZE                    = 999999  // Maximum expression size used as sentinel value
)

// StringCondition converts an L8Condition into its string representation,
// wrapping the condition in parentheses. This is useful for reconstructing
// the original query condition or for debugging purposes.
func StringCondition(this *l8api.L8Condition) string {
	buff := &bytes.Buffer{}
	buff.WriteString("(")
	toString(this, buff)
	buff.WriteString(")")
	return buff.String()
}

// VisualizeCondition creates a human-readable, indented visualization of an L8Condition
// structure. The lvl parameter controls the indentation level for nested conditions.
// This is primarily useful for debugging and understanding the parsed condition tree.
func VisualizeCondition(this *l8api.L8Condition, lvl int) string {
	buff := &bytes.Buffer{}
	buff.WriteString(space(lvl))
	buff.WriteString("Condition\n")
	if this.Comparator != nil {
		buff.WriteString(VisualizeComparator(this.Comparator, lvl+1))
	}
	if this.Next != nil {
		buff.WriteString(space(lvl))
		buff.WriteString(strings.TrimSpace(this.Oper))
		buff.WriteString("\n")
		buff.WriteString(VisualizeCondition(this.Next, lvl))
	}
	return buff.String()
}

// toString is an internal helper that recursively writes the string representation
// of a condition chain to the provided buffer.
func toString(this *l8api.L8Condition, buff *bytes.Buffer) {
	if this.Comparator != nil {
		buff.WriteString(StringComparator(this.Comparator))
	}
	if this.Next != nil {
		buff.WriteString(this.Oper)
		toString(this.Next, buff)
	}
}

// NewCondition parses a WHERE clause string into an L8Condition structure.
// The string may contain multiple comparisons connected by AND/OR operators.
// Comparisons are parsed left-to-right and linked together in a chain.
// Returns an error if the condition string contains invalid syntax.
func NewCondition(ws string) (*l8api.L8Condition, error) {
	wsLower := strings.ToLower(ws)
	loc := MAX_EXPRESSION_SIZE
	var op ConditionOperation
	and := strings.Index(wsLower, string(And))
	if and != -1 {
		loc = and
		op = And
	}
	or := strings.Index(wsLower, string(Or))
	if or != -1 && or < loc {
		loc = or
		op = Or
	}

	condition := &l8api.L8Condition{}
	if loc == MAX_EXPRESSION_SIZE {
		cmpr, e := NewCompare(ws)
		if e != nil {
			return nil, e
		}
		condition.Comparator = cmpr
		return condition, nil
	}

	cmpr, e := NewCompare(ws[0:loc])
	if e != nil {
		return nil, e
	}

	condition.Comparator = cmpr
	condition.Oper = string(op)

	ws = ws[loc+len(op):]
	next, e := NewCondition(ws)
	if e != nil {
		return nil, e
	}

	condition.Next = next
	return condition, nil
}

// getLastConditionOp finds the last occurrence of an AND or OR operator in the string.
// Returns the operator type, its position, and nil error if found.
// Returns an error if no operator is found.
func getLastConditionOp(ws string) (ConditionOperation, int, error) {
	wsLower := strings.ToLower(ws)
	loc := -1
	var op ConditionOperation

	and := strings.LastIndex(wsLower, string(And))
	if and > loc {
		op = And
		loc = and
	}

	or := strings.LastIndex(wsLower, string(Or))
	if or > loc {
		op = Or
		loc = or
	}

	if loc == -1 {
		return "", 0, errors.New("No last condition was found.")
	}
	return op, loc, nil
}

// getFirstConditionOp finds the first occurrence of an AND or OR operator in the string.
// Returns the operator type, its position, and nil error if found.
// Returns an error if no operator is found.
func getFirstConditionOp(ws string) (ConditionOperation, int, error) {
	wsLower := strings.ToLower(ws)
	loc := MAX_EXPRESSION_SIZE
	var op ConditionOperation
	and := strings.Index(wsLower, string(And))
	if and != -1 {
		loc = and
		op = And
	}
	or := strings.Index(wsLower, string(Or))
	if or != -1 && or < loc {
		loc = or
		op = Or
	}

	if loc == MAX_EXPRESSION_SIZE {
		return "", 0, errors.New("No first condition was found.")
	}

	return op, loc, nil
}
