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

// Package interpreter provides functionality for executing L8QL queries against Go data structures.
// It takes parsed query objects from the parser package and evaluates them against actual data,
// supporting filtering, sorting, pagination, and column selection.
//
// The interpreter uses reflection to access object properties and supports various comparison
// operators for different data types. It also implements the IQuery interface for integration
// with other parts of the Layer 8 ecosystem.
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

// Query is the interpreted query that can be executed against data objects.
// It holds all the parsed and resolved components needed for query execution,
// including the root type, property mappings, filter expressions, and query options.
type Query struct {
	rootType       *l8reflect.L8Node        // The root type node from the introspector
	propertiesMap  map[string]ifs.IProperty // Map of property names to property accessors
	properties     []ifs.IProperty          // Ordered list of selected properties
	where          *Expression              // The WHERE clause expression for filtering
	sortBy         string                   // Property name to sort results by
	sortByProperty *properties.Property     // Resolved sort property accessor
	descending     bool                     // Sort in descending order if true
	limit          int32                    // Maximum number of results
	page           int32                    // Page number for pagination
	matchCase      bool                     // Case-sensitive matching if true
	resources      ifs.IResources           // Resources for logging and introspection
	query          *l8api.L8Query           // The original parsed query
}

// NewFromQuery creates a new interpreted Query from a parsed L8Query protobuf message.
// It resolves all property references, creates the expression tree, and validates
// that all referenced types and properties exist. Returns an error if validation fails.
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
		sortByProperty, er := properties.PropertyOf(rootTable.TypeName+"."+iQuery.sortBy, resources)
		if er != nil {
			return nil, errors.New(er.Error())
		}
		iQuery.sortByProperty = sortByProperty
	}

	return iQuery, nil
}

// NewQuery parses an L8QL query string and creates a new interpreted Query.
// This is a convenience function that combines parsing and interpretation.
func NewQuery(gsql string, resources ifs.IResources) (*Query, error) {
	pQuery, err := parser.NewQuery(gsql, resources.Logger())
	if err != nil {
		return nil, err
	}
	return NewFromQuery(pQuery.Query(), resources)
}

// Query returns the underlying L8Query protobuf message.
func (this *Query) Query() *l8api.L8Query {
	return this.query
}

// String returns a reconstructed query string representation.
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

// RootType returns the L8Node representing the root type being queried.
func (this *Query) RootType() *l8reflect.L8Node {
	return this.rootType
}

// PropertiesMap returns a map of column names to their property accessors.
func (this *Query) PropertiesMap() map[string]ifs.IProperty {
	return this.propertiesMap
}

// Properties returns the ordered list of selected property accessors.
func (this *Query) Properties() []ifs.IProperty {
	return this.properties
}

// OnlyTopLevel returns true, indicating the query only operates on top-level objects.
func (this *Query) OnlyTopLevel() bool {
	return true
}

// Descending returns true if results should be sorted in descending order.
func (this *Query) Descending() bool {
	return this.descending
}

// MatchCase returns true if string comparisons should be case-sensitive.
func (this *Query) MatchCase() bool {
	return this.matchCase
}

// Page returns the page number for paginated results.
func (this *Query) Page() int32 {
	return this.page
}

// Limit returns the maximum number of results to return.
func (this *Query) Limit() int32 {
	return this.limit
}

// SortBy returns the property name used for sorting results.
func (this *Query) SortBy() string {
	return this.sortBy
}

// initTables resolves the root type from the query's RootType field.
func (this *Query) initTables(query *l8api.L8Query) error {
	node, ok := this.resources.Introspector().Node(query.RootType)
	if !ok {
		return this.resources.Logger().Error("Cannot find node for table ", query.RootType)
	}
	this.rootType = node
	return nil
}

// initColumns resolves the SELECT columns to property accessors.
// If the query selects "*", no specific properties are initialized.
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

// propertyPath constructs a fully qualified property path by prepending
// the root table name if not already present.
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

// match evaluates whether the given object matches the query's WHERE clause.
// Returns true if there is no WHERE clause or if the object matches.
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

// Filter applies the query's WHERE clause to a list of objects and returns
// only the matching objects. If onlySelectedColumns is true and specific
// columns were selected, returns cloned objects with only those columns populated.
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

// Match evaluates whether the given object matches the query's WHERE clause.
// This is a convenience method that logs errors and returns the boolean result.
func (this *Query) Match(any interface{}) bool {
	m, e := this.match(any)
	if e != nil {
		this.resources.Logger().Error(e)
	}
	return m
}

// SortByValue extracts the sort-by property value from the given object.
// Returns nil if no sort-by property is configured.
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

// cloneOnlyWithColumns creates a new instance of the object type and copies
// only the selected column values from the source object.
func (this *Query) cloneOnlyWithColumns(any interface{}) interface{} {
	typ := reflect.ValueOf(any).Elem().Type()
	clone := reflect.New(typ).Interface()
	for _, column := range this.properties {
		v, _ := column.Get(any)
		column.Set(clone, v)
	}
	return clone
}

// Criteria returns the WHERE clause expression for the query.
func (this *Query) Criteria() ifs.IExpression {
	return this.where
}

// KeyOf extracts a key value from the WHERE clause if one exists.
// This is used for optimization when querying by key.
func (this *Query) KeyOf() string {
	if this.where == nil {
		return ""
	}
	return this.where.keyOf()
}

// Text returns the original query text string.
func (this *Query) Text() string {
	return this.query.Text
}

// MapReduce returns true if map-reduce mode is enabled for this query.
func (this *Query) MapReduce() bool {
	return this.query.MapReduce
}

// Hash returns an MD5 hash of the query for caching and deduplication.
// The hash is based on the root type, WHERE clause, sort-by, and descending fields.
func (this *Query) Hash() string {
	h := md5.New()
	h.Write([]byte(this.Text()))
	return hex.EncodeToString(h.Sum(nil))
}

// ValueForParameter searches the WHERE clause for a comparator that references
// the given parameter name and returns the corresponding value.
// This is useful for extracting specific filter values from the query.
func (this *Query) ValueForParameter(name string) string {
	if this.where == nil {
		return ""
	}
	return this.where.ValueForParameter(name)
}
