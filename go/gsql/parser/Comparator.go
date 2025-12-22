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

// ComparatorOperation represents the comparison operators used in WHERE clause conditions.
type ComparatorOperation string

// Comparison operators supported in L8QL WHERE clauses.
const (
	Eq    ComparatorOperation = "="         // Equal comparison
	Neq   ComparatorOperation = "!="        // Not equal comparison
	GT    ComparatorOperation = ">"         // Greater than comparison
	LT    ComparatorOperation = "<"         // Less than comparison
	GTEQ  ComparatorOperation = ">="        // Greater than or equal comparison
	LTEQ  ComparatorOperation = "<="        // Less than or equal comparison
	IN    ComparatorOperation = " in "      // Membership test (value in list)
	NOTIN ComparatorOperation = " not in "  // Negative membership test (value not in list)
)

// comparators holds the ordered list of comparison operators for parsing.
// The order is important: multi-character operators must come before single-character
// ones to ensure correct matching (e.g., ">=" before ">").
var comparators = make([]ComparatorOperation, 0)

// initComparators initializes the comparators slice with all supported operators
// in the correct order for parsing. Called once on first use.
func initComparators() {
	if len(comparators) == 0 {
		comparators = append(comparators, GTEQ)
		comparators = append(comparators, LTEQ)
		comparators = append(comparators, Neq)
		comparators = append(comparators, Eq)
		comparators = append(comparators, GT)
		comparators = append(comparators, LT)
		comparators = append(comparators, NOTIN)
		comparators = append(comparators, IN)
	}
}

// StringComparator converts an L8Comparator into its string representation
// by concatenating the left operand, operator, and right operand.
func StringComparator(this *l8api.L8Comparator) string {
	buff := bytes.Buffer{}
	buff.WriteString(this.Left)
	buff.WriteString(this.Oper)
	buff.WriteString(this.Right)
	return buff.String()
}

// VisualizeComparator creates a human-readable, indented representation of an L8Comparator.
// The lvl parameter controls the indentation level.
func VisualizeComparator(this *l8api.L8Comparator, lvl int) string {
	buff := bytes.Buffer{}
	buff.WriteString(space(lvl))
	buff.WriteString("Comparator (")
	buff.WriteString(this.Left)
	buff.WriteString(string(this.Oper))
	buff.WriteString(this.Right)
	buff.WriteString(")\n")
	return buff.String()
}

// NewCompare parses a comparison expression string (e.g., "age>18", "name='John'")
// and creates an L8Comparator. It tries each comparator operator in order until
// one is found. Returns an error if no valid comparator is found or if the
// operands contain illegal characters (brackets).
func NewCompare(ws string) (*l8api.L8Comparator, error) {
	for _, op := range comparators {
		loc := strings.Index(ws, string(op))
		if loc != -1 {
			cmp := &l8api.L8Comparator{}
			cmp.Left = strings.TrimSpace(strings.ToLower(ws[0:loc]))
			cmp.Right = stripQuotes(strings.TrimSpace(ws[loc+len(op):]))
			cmp.Oper = string(op)
			if validateValue(cmp.Left) != "" {
				return nil, errors.New(validateValue(cmp.Left))
			}
			if validateValue(cmp.Right) != "" {
				return nil, errors.New(validateValue(cmp.Right))
			}
			return cmp, nil
		}
	}
	return nil, errors.New("Cannot find comparator operation in: " + ws)
}

// validateValue checks if a comparator operand contains illegal bracket characters.
// Returns an error message if brackets are found, empty string otherwise.
func validateValue(ws string) string {
	bo := strings.Index(ws, "(")
	be := strings.Index(ws, ")")
	if bo != -1 || be != -1 {
		return "Value " + ws + " contain illegale brackets."
	}
	return ""
}

// stripQuotes removes surrounding double quotes from a string value.
// If the string is not quoted, it is converted to lowercase.
// This allows for case-sensitive matching when values are explicitly quoted.
func stripQuotes(s string) string {
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		return s[1 : len(s)-1]
	}
	return strings.ToLower(s)
}
