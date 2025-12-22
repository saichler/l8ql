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
package interpreter

import (
	"bytes"
	"errors"

	"github.com/saichler/l8ql/go/gsql/parser"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
	"github.com/saichler/l8types/go/types/l8reflect"
)

// Expression represents an interpreted WHERE clause expression that can be evaluated
// against data objects. Expressions form a tree structure with conditions, child
// expressions (for grouped/parenthesized expressions), and next expressions (for chained conditions).
type Expression struct {
	condition *Condition                 // The condition at this node (leaf expression)
	operation parser.ConditionOperation  // AND/OR operator connecting to next expression
	next      *Expression                // Next expression in the chain
	child     *Expression                // Child expression (for parenthesized groups)
}

// String returns the string representation of this expression tree.
func (this *Expression) String() string {
	buff := bytes.Buffer{}
	if this.condition != nil {
		buff.WriteString(this.condition.String())
	} else {
		buff.WriteString("(")
	}
	if this.child != nil {
		buff.WriteString(this.child.String())
	}
	if this.condition == nil {
		buff.WriteString(")")
	}
	if this.next != nil {
		buff.WriteString(string(this.operation))
		buff.WriteString(this.next.String())
	}
	return buff.String()
}

// CreateExpression creates an interpreted Expression from a parsed L8Expression.
// It recursively processes the expression tree and resolves property references.
// Returns nil for nil input without error.
func CreateExpression(expr *l8api.L8Expression, rootTable *l8reflect.L8Node, resources ifs.IResources) (*Expression, error) {
	if expr == nil {
		return nil, nil
	}
	ormExpr := &Expression{}
	ormExpr.operation = parser.ConditionOperation(expr.AndOr)
	if expr.Condition != nil {
		cond, e := CreateCondition(expr.Condition, rootTable, resources)
		if e != nil {
			return nil, e
		}
		ormExpr.condition = cond
	}

	if expr.Child != nil {
		child, e := CreateExpression(expr.Child, rootTable, resources)
		if e != nil {
			return nil, e
		}
		ormExpr.child = child
	}

	if expr.Next != nil {
		next, e := CreateExpression(expr.Next, rootTable, resources)
		if e != nil {
			return nil, e
		}
		ormExpr.next = next
	}

	return ormExpr, nil
}

// Match evaluates this expression tree against the given object.
// For AND operations, all parts must match. For OR operations, any match is sufficient.
// The expression evaluates its condition, child expression, and next expression.
func (this *Expression) Match(root interface{}) (bool, error) {
	cond := true
	child := true
	next := true
	var e error
	if this.operation == parser.Or {
		cond = false
		child = false
		next = false
	}
	if this.condition != nil {
		cond, e = this.condition.Match(root)
		if e != nil {
			return false, e
		}
	}
	if this.child != nil {
		child, e = this.child.Match(root)
		if e != nil {
			return false, e
		}
	}
	if this.next != nil {
		next, e = this.next.Match(root)
		if e != nil {
			return false, e
		}
	}
	if this.operation == "" {
		return child && next && cond, nil
	}
	if this.operation == parser.And {
		return child && next && cond, nil
	}
	if this.operation == parser.Or {
		return child || next || cond, nil
	}

	return false, errors.New("Unsupported operation in match:" + string(this.operation))
}

// Condition returns the condition at this expression node.
func (this *Expression) Condition() ifs.ICondition {
	return this.condition
}

// Operator returns the AND/OR operator connecting this expression to the next.
func (this *Expression) Operator() string {
	return string(this.operation)
}

// Next returns the next expression in the chain.
func (this *Expression) Next() ifs.IExpression {
	return this.next
}

// Child returns the child expression (for parenthesized groups).
func (this *Expression) Child() ifs.IExpression {
	return this.child
}

// keyOf searches this expression tree for a literal key value.
func (this *Expression) keyOf() string {
	if this.condition != nil {
		return this.condition.keyOf()
	}
	if this.child != nil {
		return this.child.keyOf()
	}
	if this.next != nil {
		return this.next.keyOf()
	}
	return ""
}

// ValueForParameter searches this expression tree for the value associated
// with the given parameter name and returns it if found.
func (this *Expression) ValueForParameter(name string) string {
	if this.condition != nil {
		val := this.condition.ValueForParameter(name)
		if val != "" {
			return val
		}
	}
	if this.child != nil {
		val := this.child.ValueForParameter(name)
		if val != "" {
			return val
		}
	}
	if this.next != nil {
		val := this.next.ValueForParameter(name)
		if val != "" {
			return val
		}
	}
	return ""
}
