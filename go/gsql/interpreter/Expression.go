package interpreter

import (
	"bytes"
	"errors"

	"github.com/saichler/l8ql/go/gsql/parser"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
	"github.com/saichler/l8types/go/types/l8reflect"
)

type Expression struct {
	condition *Condition
	operation parser.ConditionOperation
	next      *Expression
	child     *Expression
}

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

func (this *Expression) Condition() ifs.ICondition {
	return this.condition
}

func (this *Expression) Operator() string {
	return string(this.operation)
}

func (this *Expression) Next() ifs.IExpression {
	return this.next
}

func (this *Expression) Child() ifs.IExpression {
	return this.child
}

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
