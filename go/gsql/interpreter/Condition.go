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

// Condition represents an interpreted condition that can be evaluated against data objects.
// It contains a comparator for the actual comparison and optionally a link to the next
// condition in a chain connected by AND/OR operators.
type Condition struct {
	comparator *Comparator               // The comparison to evaluate
	operation  parser.ConditionOperation // AND/OR operator connecting to next condition
	next       *Condition                // Next condition in the chain (if any)
}

// CreateCondition creates an interpreted Condition from a parsed L8Condition.
// It recursively processes linked conditions and resolves property references.
func CreateCondition(c *l8api.L8Condition, rootTable *l8reflect.L8Node, resources ifs.IResources) (*Condition, error) {
	condition := &Condition{}
	condition.operation = parser.ConditionOperation(c.Oper)
	comp, e := CreateComparator(c.Comparator, rootTable, resources)
	if e != nil {
		return nil, e
	}
	condition.comparator = comp
	if c.Next != nil {
		next, e := CreateCondition(c.Next, rootTable, resources)
		if e != nil {
			return nil, e
		}
		condition.next = next
	}
	return condition, nil
}

// String returns the string representation of this condition chain.
func (this *Condition) String() string {
	buff := &bytes.Buffer{}
	buff.WriteString("(")
	this.toString(buff)
	buff.WriteString(")")
	return buff.String()
}

// toString is a helper that recursively writes the condition chain to a buffer.
func (this *Condition) toString(buff *bytes.Buffer) {
	if this.comparator != nil {
		buff.WriteString(this.comparator.String())
	}
	if this.next != nil {
		buff.WriteString(string(this.operation))
		this.next.toString(buff)
	}
}

// Match evaluates this condition chain against the given object.
// For AND operations, all conditions must match. For OR operations, any match is sufficient.
func (this *Condition) Match(root interface{}, matchCase bool) (bool, error) {
	comp, e := this.comparator.Match(root, matchCase)
	if e != nil {
		return false, e
	}
	next := true
	if this.operation == parser.Or {
		next = false
	}
	if this.next != nil {
		next, e = this.next.Match(root, matchCase)
		if e != nil {
			return false, e
		}
	}
	if this.operation == "" {
		return next && comp, nil
	}
	if this.operation == parser.And {
		return comp && next, nil
	}
	if this.operation == parser.Or {
		return comp || next, nil
	}
	return false, errors.New("Unsupported operation in match:" + string(this.operation))
}

// Comparator returns the comparator for this condition.
func (this *Condition) Comparator() ifs.IComparator {
	return this.comparator
}

// Operator returns the operator (AND/OR) connecting this condition to the next.
func (this *Condition) Operator() string {
	return string(this.operation)
}

// Next returns the next condition in the chain, or nil if this is the last.
func (this *Condition) Next() ifs.ICondition {
	return this.next
}

// keyOf searches this condition chain for a literal key value.
func (this *Condition) keyOf() string {
	if this.comparator != nil {
		return this.comparator.keyOf()
	}
	if this.next != nil {
		return this.next.keyOf()
	}
	return ""
}

// ValueForParameter searches this condition chain for the value associated
// with the given parameter name and returns it if found.
func (this *Condition) ValueForParameter(name string) string {
	if this.comparator != nil {
		val := this.comparator.ValueForParameter(name)
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
