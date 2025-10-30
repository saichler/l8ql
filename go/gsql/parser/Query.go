package parser

import (
	"bytes"

	"github.com/saichler/l8types/go/ifs"
	"github.com/saichler/l8types/go/types/l8api"

	"strconv"
	"strings"
)

type PQuery struct {
	log    ifs.ILogger
	pquery l8api.L8Query
}

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

const (
	Select     = "select"
	From       = "from"
	Where      = "where"
	SortBy     = "sort-by"
	Descending = "descending"
	Ascending  = "ascending"
	Limit      = "limit"
	Page       = "page"
	MatchCase  = "match-case"
	MapReduce  = "mapreduce"
)

var words = []string{Select, From, Where, SortBy, Descending, Ascending, Limit, Page, MatchCase, MapReduce}

func (this *PQuery) Query() *l8api.L8Query {
	return &this.pquery
}

func NewQuery(query string, log ifs.ILogger) (*PQuery, error) {
	cwql := &PQuery{}
	cwql.pquery.Text = query
	cwql.log = log
	e := cwql.init()
	return cwql, e
}

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
	data.select_ = getSplitTag(sql, Select)
	data.from_ = getTag(sql, From)
	data.where_ = getTag(sql, Where)
	data.descending_ = getBoolTag(sql, Descending)
	data.ascending_ = getBoolTag(sql, Ascending)
	data.limit_ = getTag(sql, Limit)
	data.page_ = getTag(sql, Page)
	data.sortby_ = getTag(sql, SortBy)
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

func getTag(str, tag string) string {
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
	return strings.TrimSpace(str[index:index2])
}

func getSplitTag(str, tag string) []string {
	result := make([]string, 0)
	data := getTag(str, tag)
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
	if p.matchcase_ != "" {
		this.pquery.MatchCase = true
	}
	if p.mapreduce_ != "" {
		this.pquery.MapReduce = true
	}
	return nil
}
