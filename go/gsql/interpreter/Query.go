package interpreter

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"errors"
	"reflect"
	"strings"

	"github.com/saichler/l8ql/go/gsql/parser"
	"github.com/saichler/l8reflect/go/reflect/properties"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"
	"github.com/saichler/l8types/go/types/l8reflect"
)

type Query struct {
	rootType       *l8reflect.L8Node
	propertiesMap  map[string]ifs.IProperty
	properties     []ifs.IProperty
	where          *Expression
	sortBy         string
	sortByProperty *properties.Property
	descending     bool
	limit          int32
	page           int32
	matchCase      bool
	resources      ifs.IResources
	query          *l8api.L8Query
}

func NewFromQuery(query *l8api.L8Query, resources ifs.IResources) (*Query, error) {
	iQuery := &Query{}
	iQuery.propertiesMap = make(map[string]ifs.IProperty)
	iQuery.properties = make([]ifs.IProperty, 0)
	iQuery.descending = query.Descending
	iQuery.matchCase = query.MatchCase
	iQuery.page = query.Page
	iQuery.limit = query.Limit
	iQuery.sortBy = query.SortBy
	iQuery.resources = resources
	iQuery.query = query

	err := iQuery.initTables(query)
	if err != nil {
		return nil, err
	}

	err = iQuery.initColumns(query, resources)
	if err != nil {
		return nil, err
	}

	rootTable := iQuery.RootType()
	if rootTable == nil {
		return nil, errors.New("root table is nil")
	}

	expr, err := CreateExpression(query.Criteria, rootTable, resources)
	if err != nil {
		return nil, err
	}
	iQuery.where = expr

	if iQuery.sortBy != "" {
		sortByProperty, er := properties.PropertyOf(iQuery.sortBy, resources)
		if er != nil {
			return nil, errors.New(er.Error())
		}
		iQuery.sortByProperty = sortByProperty
	}

	return iQuery, nil
}

func NewQuery(gsql string, resources ifs.IResources) (*Query, error) {
	pQuery, err := parser.NewQuery(gsql, resources.Logger())
	if err != nil {
		return nil, err
	}
	return NewFromQuery(pQuery.Query(), resources)
}

func (this *Query) Query() *l8api.L8Query {
	return this.query
}

func (this *Query) String() string {
	buff := bytes.Buffer{}
	buff.WriteString("Select ")
	first := true

	for _, column := range this.Properties() {
		if !first {
			buff.WriteString(", ")
		}
		id, _ := column.PropertyId()
		buff.WriteString(id)
		first = false
	}

	buff.WriteString(" From ")
	buff.WriteString(this.rootType.TypeName)

	if this.where != nil {
		buff.WriteString(" Where ")
		buff.WriteString(this.where.String())
	}
	return buff.String()
}

func (this *Query) RootType() *l8reflect.L8Node {
	return this.rootType
}

func (this *Query) PropertiesMap() map[string]ifs.IProperty {
	return this.propertiesMap
}

func (this *Query) Properties() []ifs.IProperty {
	return this.properties
}

func (this *Query) OnlyTopLevel() bool {
	return true
}

func (this *Query) Descending() bool {
	return this.descending
}

func (this *Query) MatchCase() bool {
	return this.matchCase
}

func (this *Query) Page() int32 {
	return this.page
}

func (this *Query) Limit() int32 {
	return this.limit
}

func (this *Query) SortBy() string {
	return this.sortBy
}

func (this *Query) initTables(query *l8api.L8Query) error {
	node, ok := this.resources.Introspector().Node(query.RootType)
	if !ok {
		return this.resources.Logger().Error("Cannot find node for table ", query.RootType)
	}
	this.rootType = node
	return nil
}

func (this *Query) initColumns(query *l8api.L8Query, resources ifs.IResources) error {
	if query.Properties != nil && len(query.Properties) == 1 && query.Properties[0] == "*" {
		return nil
	} else {
		for _, col := range query.Properties {
			propPath := propertyPath(col, this.rootType.TypeName)
			prop, err := properties.PropertyOf(propPath, resources)
			if err != nil {
				return this.resources.Logger().Error("cannot find property for col ", propPath, ":", err.Error())
			}
			this.propertiesMap[col] = prop
			this.properties = append(this.properties, prop)
		}
	}
	return nil
}

func propertyPath(colName, rootTable string) string {
	rootTable = strings.ToLower(rootTable)
	if strings.Contains(colName, rootTable) {
		return colName
	}
	buff := bytes.Buffer{}
	buff.WriteString(rootTable)
	buff.WriteString(".")
	buff.WriteString(colName)
	return buff.String()
}

func (this *Query) match(root interface{}) (bool, error) {
	if root == nil {
		return false, nil
	}
	if this.rootType == nil {
		return false, nil
	}
	if this.where == nil {
		return true, nil
	}
	return this.where.Match(root)
}

func (this *Query) Filter(list []interface{}, onlySelectedColumns bool) []interface{} {
	result := make([]interface{}, 0)
	for _, i := range list {
		if this.Match(i) {
			if !onlySelectedColumns || len(this.properties) == 0 {
				result = append(result, i)
			} else {
				result = append(result, this.cloneOnlyWithColumns(i))
			}
		}
	}
	return result
}

func (this *Query) Match(any interface{}) bool {
	m, e := this.match(any)
	if e != nil {
		this.resources.Logger().Error(e)
	}
	return m
}

func (this *Query) SortByValue(v interface{}) interface{} {
	if this.sortBy == "" {
		return nil
	}
	resp, e := this.sortByProperty.Get(v)
	if e != nil {
		this.resources.Logger().Error(e)
	}
	return resp
}

func (this *Query) cloneOnlyWithColumns(any interface{}) interface{} {
	typ := reflect.ValueOf(any).Elem().Type()
	clone := reflect.New(typ).Interface()
	for _, column := range this.properties {
		v, _ := column.Get(any)
		column.Set(clone, v)
	}
	return clone
}

func (this *Query) Criteria() ifs.IExpression {
	return this.where
}

func (this *Query) KeyOf() string {
	if this.where == nil {
		return ""
	}
	return this.where.keyOf()
}

func (this *Query) Text() string {
	return this.query.Text
}

func (this *Query) Hash() string {
	buff := bytes.Buffer{}
	if this.rootType != nil {
		buff.WriteString(this.rootType.TypeName)
	}
	if this.where != nil {
		buff.WriteString(this.where.String())
	}
	buff.WriteString(this.sortBy)
	h := md5.New()
	h.Write(buff.Bytes())
	return hex.EncodeToString(h.Sum(nil))
}

func (this *Query) ValueForParameter(name string) string {
	if this.where == nil {
		return ""
	}
	return this.where.ValueForParameter(name)
}
