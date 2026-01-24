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

// StringExpression converts an L8Expression tree into its string representation.
// It recursively processes child expressions and concatenates them with their
// AND/OR operators to recreate the original expression string.
func StringExpression(this *l8api.L8Expression) string {
	buff := bytes.Buffer{}
	if this.Condition != nil {
		buff.WriteString(StringCondition(this.Condition))
	} else {
		buff.WriteString("(")
	}
	if this.Child != nil {
		buff.WriteString(StringExpression(this.Child))
	}
	if this.Condition == nil {
		buff.WriteString(")")
	}
	if this.Next != nil {
		buff.WriteString(this.AndOr)
		buff.WriteString(StringExpression(this.Next))
	}
	return buff.String()
}

// VisualizeExpression creates a human-readable, indented visualization of an L8Expression tree.
// The lvl parameter controls the indentation level for nested expressions.
// This is primarily useful for debugging and understanding the parsed expression structure.
func VisualizeExpression(this *l8api.L8Expression, lvl int) string {
	buff := bytes.Buffer{}
	buff.WriteString(space(lvl))
	buff.WriteString("Expression\n")
	if this.Condition != nil {
		buff.WriteString(VisualizeCondition(this.Condition, lvl+1))
	}
	if this.Child != nil {
		buff.WriteString(VisualizeExpression(this.Child, lvl+1))
	}
	if this.Next != nil {
		buff.WriteString(space(lvl))
		buff.WriteString(strings.TrimSpace(this.AndOr))
		buff.WriteString("\n")
		buff.WriteString(VisualizeExpression(this.Next, lvl))
	}
	return buff.String()
}

// space generates an indentation string with a pipe prefix and dashes
// based on the nesting level. Used by visualization functions.
func space(lvl int) string {
	buff := bytes.Buffer{}
	buff.WriteString("|")
	for i := 0; i < lvl; i++ {
		buff.WriteString("--")
	}
	return buff.String()
}

// parseExpression is the main entry point for parsing WHERE clause expressions.
// It handles expressions with or without parentheses and delegates to specialized
// parsing functions based on the structure of the expression.
func parseExpression(ws string) (*l8api.L8Expression, error) {
	ws = strings.TrimSpace(ws)
	bo := getBO(ws)
	if bo == -1 {
		return parseNoBrackets(ws)
	}

	if bo > 0 {
		return parseBeforeBrackets(ws, bo)
	}

	return parseWithBrackets(ws, bo)
}

// parseWithBrackets handles expressions that start with an opening parenthesis.
// It recursively parses the content within the brackets and any following expressions.
func parseWithBrackets(ws string, bo int) (*l8api.L8Expression, error) {
	be, e := getBE(ws, bo)
	if e != nil {
		return nil, e
	}
	expr := &l8api.L8Expression{}
	child, e := parseExpression(ws[1:be])
	if e != nil {
		return nil, e
	}

	expr.Child = child

	if be < len(ws)-1 {
		op, loc, e := getFirstConditionOp(ws[be+1:])
		if e != nil {
			return nil, e
		}
		expr.AndOr = string(op)
		next, e := parseExpression(ws[be+1+loc+len(op):])
		if e != nil {
			return nil, e
		}
		expr.Next = next
	}
	return expr, nil
}

// parseBeforeBrackets handles expressions where content appears before the first opening parenthesis.
// It parses the prefix conditions and links them to the bracketed expression that follows.
func parseBeforeBrackets(ws string, bo int) (*l8api.L8Expression, error) {
	prefix := ws[0:bo]
	op, loc, e := getLastConditionOp(prefix)
	if e != nil {
		return nil, e
	}
	expr, e := parseNoBrackets(prefix[0:loc])
	if e != nil {
		return nil, e
	}
	expr.AndOr = string(op)
	next, e := parseExpression(ws[bo:])
	if e != nil {
		return nil, e
	}
	expr.Next = next
	return expr, nil
}

// parseNoBrackets handles simple expressions without parentheses.
// It creates an expression containing just a condition chain.
func parseNoBrackets(ws string) (*l8api.L8Expression, error) {
	expr := &l8api.L8Expression{}
	condition, e := NewCondition(ws)
	if e != nil {
		return nil, e
	}
	expr.Condition = condition
	return expr, nil
}

// getBO (get Bracket Open) finds the position of the first opening parenthesis in the string.
// Returns -1 if no opening parenthesis is found.
func getBO(ws string) int {
	return strings.Index(ws, "(")
}

// getBE (get Bracket End) finds the matching closing parenthesis for an opening parenthesis.
// It counts nested brackets to find the correct matching close bracket.
// Returns an error if no matching closing bracket is found.
func getBE(ws string, bo int) (int, error) {
	count := 0
	for i := bo; i < len(ws); i++ {
		if byte(ws[i]) == byte('(') {
			count++
		} else if byte(ws[i]) == byte(')') {
			count--
		}
		if count == 0 {
			return i, nil
		}
	}
	return -1, errors.New("Missing close bracket in: " + ws)
}
