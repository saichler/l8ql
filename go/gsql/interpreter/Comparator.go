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
	"github.com/saichler/l8ql/go/gsql/interpreter/comparators"
	"github.com/saichler/l8ql/go/gsql/parser"
	"github.com/saichler/l8reflect/go/reflect/properties"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
	"github.com/saichler/l8types/go/types/l8reflect"
	"reflect"
	"strings"
)

// Comparator represents an interpreted comparison that can be evaluated against data objects.
// It holds the left and right operands (either literal values or property references)
// and the comparison operation to perform.
type Comparator struct {
	left          string                     // Left operand as string (property name or literal)
	leftProperty  *properties.Property       // Resolved property for left operand (if applicable)
	operation     parser.ComparatorOperation // The comparison operation (=, !=, >, <, etc.)
	right         string                     // Right operand as string (property name or literal)
	rightProperty *properties.Property       // Resolved property for right operand (if applicable)
}

// Comparable is the interface implemented by comparison operators.
// Each comparator operation has a corresponding Comparable implementation.
type Comparable interface {
	Compare(interface{}, interface{}) bool
}

// comparables maps comparison operations to their implementations.
var comparables = make(map[parser.ComparatorOperation]Comparable)

// init initializes the comparables map with all supported comparison implementations.
func init() {
	comparables[parser.Eq] = comparators.NewEqual()
	comparables[parser.Neq] = comparators.NewNotEqual()
	comparables[parser.NOTIN] = comparators.NewNotIN()
	comparables[parser.IN] = comparators.NewIN()
	comparables[parser.GT] = comparators.NewGreaterThan()
	comparables[parser.LT] = comparators.NewLessThan()
	comparables[parser.GTEQ] = comparators.NewGreaterThanOrEqual()
	comparables[parser.LTEQ] = comparators.NewLessThanOrEqual()
}

// String returns the string representation of this comparator.
func (this *Comparator) String() string {
	buff := bytes.Buffer{}
	if this.leftProperty != nil {
		pid, _ := this.leftProperty.PropertyId()
		buff.WriteString(pid)
	} else {
		buff.WriteString(this.left)
	}
	buff.WriteString(string(this.operation))
	if this.rightProperty != nil {
		pid, _ := this.rightProperty.PropertyId()
		buff.WriteString(pid)
	} else {
		buff.WriteString(this.right)
	}
	return buff.String()
}

// CreateComparator creates an interpreted Comparator from a parsed L8Comparator.
// It attempts to resolve both operands as property references; at least one must resolve.
// Returns an error if neither operand can be resolved to a property.
func CreateComparator(c *l8api.L8Comparator, rootTable *l8reflect.L8Node, resources ifs.IResources) (*Comparator, error) {
	ormComp := &Comparator{}
	ormComp.operation = parser.ComparatorOperation(c.Oper)
	ormComp.left = c.Left
	ormComp.right = c.Right
	leftProp := propertyPath(ormComp.left, rootTable.TypeName)
	rightProp := propertyPath(ormComp.right, rootTable.TypeName)
	ormComp.leftProperty, _ = properties.PropertyOf(leftProp, resources)
	ormComp.rightProperty, _ = properties.PropertyOf(rightProp, resources)
	if ormComp.leftProperty == nil && ormComp.rightProperty == nil {
		return nil, errors.New("No Field was found for comparator: " + c.String())
	}
	return ormComp, nil
}

// Match evaluates this comparison against the given object.
// It retrieves the property values and delegates to the appropriate Comparable implementation.
func (this *Comparator) Match(root interface{}, matchCase bool) (bool, error) {
	var leftValue interface{}
	var rightValue interface{}
	var err error
	if this.leftProperty != nil {
		leftValue, err = this.leftProperty.Get(root)
		if err != nil {
			return false, err
		}
	} else {
		leftValue = this.left
	}
	if this.rightProperty != nil {
		rightValue, err = this.rightProperty.Get(root)
		return false, err
	} else {
		rightValue = this.right
	}
	if !matchCase {
		leftValue = toLowerValue(leftValue)
		rightValue = toLowerValue(rightValue)
	}
	matcher := comparables[this.operation]
	if matcher == nil {
		panic("No Matcher for: " + this.operation + " operation.")
	}
	return matcher.Compare(leftValue, rightValue), nil
}

// Left returns the left operand as a string.
func (this *Comparator) Left() string {
	return this.left
}

// LeftProperty returns the resolved property for the left operand, or nil if it's a literal.
func (this *Comparator) LeftProperty() ifs.IProperty {
	return this.leftProperty
}

// Right returns the right operand as a string.
func (this *Comparator) Right() string {
	return this.right
}

// RightProperty returns the resolved property for the right operand, or nil if it's a literal.
func (this *Comparator) RightProperty() ifs.IProperty {
	return this.rightProperty
}

// Operator returns the comparison operator as a string.
func (this *Comparator) Operator() string {
	return string(this.operation)
}

// keyOf returns the literal operand value if one side is a literal and the other is a property.
func (this *Comparator) keyOf() string {
	if this.leftProperty == nil {
		return this.left
	}
	if this.rightProperty == nil {
		return this.right
	}
	return ""
}

// ValueForParameter returns the value paired with the given parameter name.
// If the right operand matches the name, returns the left value, and vice versa.
func (this *Comparator) ValueForParameter(name string) string {
	if this.right == name {
		return this.left
	}
	if this.left == name {
		return this.right
	}
	return ""
}

// toLowerValue converts string values to lowercase for case-insensitive comparison.
// It handles plain strings, slices of strings, and maps with string values.
func toLowerValue(value interface{}) interface{} {
	if s, ok := value.(string); ok {
		return strings.ToLower(s)
	}

	v := reflect.ValueOf(value)
	if !v.IsValid() {
		return value
	}

	switch v.Kind() {
	case reflect.Slice:
		if v.Len() == 0 {
			return value
		}
		elem := v.Index(0)
		elemKind := elem.Kind()
		if elemKind == reflect.Interface {
			elemKind = reflect.ValueOf(elem.Interface()).Kind()
		}
		if elemKind == reflect.String {
			result := make([]interface{}, v.Len())
			for i := 0; i < v.Len(); i++ {
				e := v.Index(i)
				if e.Kind() == reflect.Interface {
					e = reflect.ValueOf(e.Interface())
				}
				result[i] = strings.ToLower(e.String())
			}
			return result
		}
	case reflect.Map:
		mapType := v.Type()
		elemKind := mapType.Elem().Kind()
		if elemKind == reflect.Interface {
			for _, key := range v.MapKeys() {
				val := v.MapIndex(key)
				if reflect.ValueOf(val.Interface()).Kind() == reflect.String {
					elemKind = reflect.String
					break
				}
			}
		}
		if elemKind == reflect.String {
			result := reflect.MakeMap(mapType)
			for _, key := range v.MapKeys() {
				val := v.MapIndex(key)
				if val.Kind() == reflect.Interface {
					val = reflect.ValueOf(val.Interface())
				}
				result.SetMapIndex(key, reflect.ValueOf(strings.ToLower(val.String())))
			}
			return result.Interface()
		}
	}

	return value
}
