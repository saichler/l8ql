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

// Package parser provides functionality for parsing L8QL (Layer 8 Query Language) query strings
// into structured protobuf message objects. L8QL is a SQL-like query language designed for
// querying Go data structures.
//
// The parser supports the following clauses:
//   - SELECT: Specify which properties/columns to retrieve (comma-separated)
//   - FROM: Specify the root type to query
//   - WHERE: Filter conditions with comparators (=, !=, >, <, >=, <=, in, not in)
//   - SORT-BY: Property to sort results by
//   - DESCENDING/ASCENDING: Sort order modifiers
//   - LIMIT: Maximum number of results (up to 1000)
//   - PAGE: Pagination offset
//   - MATCH-CASE: Enable case-sensitive matching
//   - MAPREDUCE: Enable map-reduce mode
//
// Example query:
//
//	"select name,age from Person where age>18 and status='active' sort-by name limit 10"
package parser

import (
	"bytes"
	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"

	"strconv"
	"strings"
)

// PQuery is the parsed query wrapper that holds the parsed L8Query protobuf message
// along with a logger for error reporting during parsing.
type PQuery struct {
	log    ifs.ILogger
	pquery l8api.L8Query
}

// parsed is an internal struct that holds the raw parsed values extracted from
// the query string before they are converted to their final typed representations.
type parsed struct {
	select_     []string
	from_       string
	where_      string
	sortby_     string
	descending_ string
	ascending_  string
	limit_      string
	page_       string
	matchcase_  string
	mapreduce_  string
}

// Query clause keywords used for parsing L8QL query strings.
const (
	Select     = "select"     // SELECT clause keyword for specifying properties to retrieve
	From       = "from"       // FROM clause keyword for specifying the root type
	Where      = "where"      // WHERE clause keyword for filter conditions
	SortBy     = "sort-by"    // SORT-BY clause keyword for ordering results
	Descending = "descending" // DESCENDING keyword for descending sort order
	Ascending  = "ascending"  // ASCENDING keyword for ascending sort order
	Limit      = "limit"      // LIMIT clause keyword for maximum results
	Page       = "page"       // PAGE clause keyword for pagination
	MatchCase  = "match-case" // MATCH-CASE keyword for case-sensitive matching
	MapReduce  = "mapreduce"  // MAPREDUCE keyword for enabling map-reduce mode
)

// words contains all query keywords used for parsing clause boundaries.
var words = []string{Select, From, Where, SortBy, Descending, Ascending, Limit, Page, MatchCase, MapReduce}

// Query returns a pointer to the underlying L8Query protobuf message
// that was parsed from the query string.
func (this *PQuery) Query() *l8api.L8Query {
	return &this.pquery
}

// NewQuery parses an L8QL query string and returns a new PQuery instance.
// The query string should follow L8QL syntax with clauses like SELECT, FROM, WHERE, etc.
// Returns an error if the query string contains invalid syntax or values.
func NewQuery(query string, log ifs.ILogger) (*PQuery, error) {
	cwql := &PQuery{}
	cwql.pquery.Text = query
	cwql.log = log
	e := cwql.init()
	return cwql, e
}

// TrimAndLowerNoKeys trims whitespace and converts the query string to lowercase,
// but preserves the original case for content within square brackets (keys).
// This allows for case-insensitive keyword matching while maintaining case-sensitive key values.
func TrimAndLowerNoKeys(sql string) string {
	buff := bytes.Buffer{}
	sql = strings.TrimSpace(sql)
	keyOpen := false
	for _, c := range sql {
		if c == '[' {
			keyOpen = true
		} else if c == ']' {
			keyOpen = false
		}
		if !keyOpen {
			buff.WriteString(strings.ToLower(string(c)))
		} else {
			buff.WriteString(string(c))
		}
	}
	return buff.String()
}

func (this *PQuery) split() *parsed {
	sql := TrimAndLowerNoKeys(this.pquery.Text)
	data := &parsed{}
	data.select_ = getSplitTag(sql, this.pquery.Text, Select)
	data.from_ = getTag(sql, this.pquery.Text, From)
	data.where_ = getTag(sql, this.pquery.Text, Where)
	data.descending_ = getBoolTag(sql, Descending)
	data.ascending_ = getBoolTag(sql, Ascending)
	data.limit_ = getTag(sql, this.pquery.Text, Limit)
	data.page_ = getTag(sql, this.pquery.Text, Page)
	data.sortby_ = getTag(sql, this.pquery.Text, SortBy)
	data.matchcase_ = getBoolTag(sql, MatchCase)
	data.mapreduce_ = getBoolTag(sql, MapReduce)
	return data
}

func getBoolTag(str, tag string) string {
	index := strings.Index(str, tag)
	if index != -1 {
		return "true"
	}
	return "false"
}

func getTag(str, orig, tag string) string {
	index := strings.Index(str, tag)
	if index == -1 {
		return ""
	}
	index += len(tag)
	index2 := len(str)
	for _, t := range words {
		if t != tag {
			index3 := strings.Index(str, t)
			if index3 > index && index3 < index2 {
				index2 = index3
			}
		}
	}
	return strings.TrimSpace(orig[index:index2])
}

func getSplitTag(str, orig, tag string) []string {
	result := make([]string, 0)
	data := getTag(str, orig, tag)
	if data == "" {
		return result
	}
	split := strings.Split(data, ",")
	for _, t := range split {
		result = append(result, t)
	}
	return result
}

func (this *PQuery) init() error {
	p := this.split()
	this.pquery.Properties = make([]string, 0)
	this.pquery.RootType = strings.TrimSpace(p.from_)
	for _, col := range p.select_ {
		this.pquery.Properties = append(this.pquery.Properties, col)
	}
	if p.where_ != "" {
		where, e := parseExpression(p.where_)
		if e != nil {
			return e
		}
		this.pquery.Criteria = where
	}
	if p.limit_ != "" {
		limit, e := strconv.Atoi(p.limit_)
		if e != nil {
			this.log.Error("Invalid limit:", p.limit_, ", setting limity to 10")
			limit = 10
		}
		if limit >= 1000 {
			return this.log.Error("Invalid limit: Limit is limited up to 1000 elements")
		}
		this.pquery.Limit = int32(limit)
	}
	if p.page_ != "" {
		page, e := strconv.Atoi(p.page_)
		if e != nil {
			return this.log.Error("Invalid page:", p.page_, ":", e.Error())
		}
		this.pquery.Page = int32(page)
	}
	this.pquery.SortBy = p.sortby_
	if p.descending_ == "true" {
		this.pquery.Descending = true
	}
	if p.ascending_ == "true" {
		this.pquery.Descending = false
	}
	if p.matchcase_ == "true" {
		this.pquery.MatchCase = true
	}
	if p.mapreduce_ == "true" {
		this.pquery.MapReduce = true
	}
	return nil
}
