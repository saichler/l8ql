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
)

type Comparator struct {
	left          string
	leftProperty  *properties.Property
	operation     parser.ComparatorOperation
	right         string
	rightProperty *properties.Property
}

type Comparable interface {
	Compare(interface{}, interface{}) bool
}

var comparables = make(map[parser.ComparatorOperation]Comparable)

func initComparables() {
	if len(comparables) == 0 {
		comparables[parser.Eq] = comparators.NewEqual()
		comparables[parser.Neq] = comparators.NewNotEqual()
		comparables[parser.NOTIN] = comparators.NewNotIN()
		comparables[parser.IN] = comparators.NewIN()
		comparables[parser.GT] = comparators.NewGreaterThan()
		comparables[parser.LT] = comparators.NewLessThan()
		comparables[parser.GTEQ] = comparators.NewGreaterThanOrEqual()
		comparables[parser.LTEQ] = comparators.NewLessThanOrEqual()
	}
}

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

func CreateComparator(c *l8api.L8Comparator, rootTable *l8reflect.L8Node, resources ifs.IResources) (*Comparator, error) {
	initComparables()
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

func (this *Comparator) Match(root interface{}) (bool, error) {
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
	matcher := comparables[this.operation]
	if matcher == nil {
		panic("No Matcher for: " + this.operation + " operation.")
	}
	return matcher.Compare(leftValue, rightValue), nil
}

func (this *Comparator) Left() string {
	return this.left
}

func (this *Comparator) LeftProperty() ifs.IProperty {
	return this.leftProperty
}

func (this *Comparator) Right() string {
	return this.right
}

func (this *Comparator) RightProperty() ifs.IProperty {
	return this.rightProperty
}

func (this *Comparator) Operator() string {
	return string(this.operation)
}

func (this *Comparator) keyOf() string {
	if this.leftProperty == nil {
		return this.left
	}
	if this.rightProperty == nil {
		return this.right
	}
	return ""
}

func (this *Comparator) ValueForParameter(name string) string {
	if this.right == name {
		return this.left
	}
	if this.left == name {
		return this.right
	}
	return ""
}
