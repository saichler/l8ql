# L8QL - Layer 8 Query Language

[![Go Version](https://img.shields.io/badge/go-1.25.4-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/license-GPL%20v3-blue.svg)](LICENSE)

L8QL is a lightweight, pure-Go, reflection-based SQL-like query language for filtering Go data structures in memory. No database, no ORM - just struct filtering via reflection. It provides a single, simple, and common API to query graph model data of any Go struct at runtime.

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Architecture](#architecture)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Query Syntax](#query-syntax)
- [Usage Examples](#usage-examples)
- [API Reference](#api-reference)
- [Testing](#testing)
- [Project Structure](#project-structure)
- [Dependencies](#dependencies)
- [Contributing](#contributing)
- [License](#license)

## Overview

We keep inventing the wheel, over and over again, trying to create APIs for our services/products and spending enormous time and money trying to integrate different products. Most software engineers consider infrastructure components like Kafka, NATS, DB, ETCD, etc., as the "Wheel" and rush to implement and use those infrastructure components. However, they are doing the complete opposite.

While creating those from scratch is a nice challenge, it isn't as expensive as maintaining APIs and integrations over time. Using infrastructure components is a very easy task that can take a month or even weeks, while building a stable API and integrating with different products might take years, huge amounts of money, and constant costly maintenance over time.

If we do an analogy to Language, the infrastructure components are the alphabet, while the API is the actual Languages. Just as two persons, each knowing a different language but with the same alphabet, cannot speak to each other, two products built with the same infrastructure cannot communicate with each other and require very expensive, highly maintenance integrations.

**L8QL comes to ease the language/API challenge by presenting a single, simple & common API to query the graph model & data of a product at runtime.**

## Features

- **SQL-like Query Language** - Familiar syntax for developers
- **Two-Stage Compilation** - Parser (string to protobuf) then Interpreter (protobuf to executable query)
- **Graph Model Support** - Native support for complex nested structures
- **Runtime Introspection** - Automatic model discovery and schema generation
- **Filter and Match** - Powerful filtering with wildcard pattern matching (`J*` matches `John`, `Jane`)
- **Aggregation** - `count(*)`, `sum()`, `avg()`, `min()`, `max()` with `group-by` and `having`
- **Sorting and Pagination** - `sort-by`, `limit`, `page`, `ascending`/`descending`
- **Case Sensitivity Control** - Optional `match-case` for case-sensitive string matching
- **Deep Path Navigation** - Query nested objects, arrays, and maps seamlessly
- **MapReduce Mode** - Built-in `mapreduce` flag for distributed query execution
- **Hash Support** - MD5 hash generation for query caching and deduplication
- **Type-Safe Comparators** - Dedicated comparators for strings, integers, unsigned integers, booleans, and pointers
- **Zero External Runtime Dependencies** - Pure Go with reflection

## Architecture

L8QL uses a two-stage compilation model:

```
Query String --> [Parser] --> L8Query Protobuf --> [Interpreter] --> Executable Query
                 (syntax)                          (semantic)
```

### 1. Parser
The parser (`go/gsql/parser/`) takes a query string and produces an `L8Query` protobuf message. It validates syntax only - no type checking. The query is divided into:
- **Selected Columns** - Fields to return (including aggregate functions)
- **Table** - Target struct type
- **Where Clause** - Expression tree of conditions and comparators
- **Modifiers** - Sort-by, limit, page, match-case, mapreduce, group-by, having

### 2. Interpreter
The interpreter (`go/gsql/interpreter/`) takes the parsed protobuf and an `IResources` instance (which provides the introspector) and produces an executable `Query`. It validates the query against the introspected type schema, resolving string field names to actual property accessors.

### 3. Comparators
Type-aware comparison operators (`go/gsql/interpreter/comparators/`) handle the specifics of comparing different Go types: strings (with wildcard support), integers, unsigned integers, booleans, and pointers/nil.

### 4. Accumulator
The accumulator (`go/gsql/interpreter/Accumulator.go`) tracks aggregate computation state for `count`, `sum`, `avg`, `min`, and `max` operations across grouped results.

## Installation

```bash
go get github.com/saichler/l8ql/go
```

## Quick Start

```go
package main

import (
    "fmt"
    "github.com/saichler/l8ql/go/gsql/interpreter"
)

type Employee struct {
    Name       string
    Age        int32
    Department string
    Salary     float64
    Addresses  []Address
}

type Address struct {
    Line1   string
    City    string
    Country string
}

func main() {
    // Create and execute a query
    query, err := interpreter.NewQuery(
        "select name, age from Employee where name='John' or age>25",
        resources, // your IResources implementation
    )
    if err != nil {
        panic(err)
    }

    employees := []interface{}{
        &Employee{Name: "John", Age: 30, Department: "Engineering"},
        &Employee{Name: "Jane", Age: 25, Department: "Sales"},
        &Employee{Name: "Bob", Age: 35, Department: "Engineering"},
    }

    // Filter matching elements
    results := query.Filter(employees, false)
    for _, r := range results {
        fmt.Printf("Matched: %+v\n", r)
    }

    // Or test individual objects
    for _, emp := range employees {
        if query.Match(emp) {
            fmt.Printf("Match: %+v\n", emp)
        }
    }
}
```

## Query Syntax

### Basic Structure
```sql
select <columns> from <type>
    [where <conditions>]
    [sort-by <column>] [ascending|descending]
    [limit <number>] [page <number>]
    [match-case]
    [mapreduce]
    [group-by <columns>] [having <conditions>]
```

### Comparison Operators
| Operator | Description |
|----------|-------------|
| `=` | Equal (supports wildcards: `name='J*'`) |
| `!=` | Not Equal |
| `>` | Greater Than |
| `>=` | Greater Than or Equal |
| `<` | Less Than |
| `<=` | Less Than or Equal |
| `in` | Value in list |
| `not in` | Value not in list |

### Logical Operators
| Operator | Description |
|----------|-------------|
| `and` | Logical AND |
| `or` | Logical OR |
| `()` | Parentheses for grouping |

### Aggregate Functions
| Function | Description |
|----------|-------------|
| `count(*)` | Count all rows |
| `count(field)` | Count non-null values |
| `sum(field)` | Sum numeric values |
| `avg(field)` | Average numeric values |
| `min(field)` | Minimum value |
| `max(field)` | Maximum value |

### Modifiers
| Modifier | Description |
|----------|-------------|
| `sort-by <col>` | Sort results by column |
| `ascending` / `descending` | Sort direction (default: ascending) |
| `limit <n>` | Limit results (max 1000) |
| `page <n>` | Page number for pagination |
| `match-case` | Enable case-sensitive string matching |
| `mapreduce` | Enable map-reduce mode for distributed execution |
| `group-by <cols>` | Group results for aggregation |
| `having <conds>` | Filter groups after aggregation |

## Usage Examples

### Basic Queries
```sql
-- Select all fields
select * from Employee

-- Select specific fields with conditions
select name, age from Employee where age > 25

-- Complex conditions with parentheses
select name from Employee where (age > 25 and department = 'Engineering') or country = 'USA'
```

### Deep Path Navigation
```sql
-- Query nested objects
select name from Employee where addresses.country = 'USA'

-- Query by map values
select name from Employee where metadata.department = 'Engineering'
```

### Pattern Matching
```sql
-- Wildcard matching
select * from Employee where name = 'J*'

-- Case-sensitive matching
select * from Employee where name = 'John' match-case
```

### Sorting and Pagination
```sql
-- Sort descending with pagination
select * from Employee where age > 25 sort-by age descending limit 10 page 1
```

### Membership Tests
```sql
-- In / Not In
select * from Employee where status in ('active', 'pending')
select * from Employee where department not in ('HR', 'Legal')
```

### Aggregation
```sql
-- Count all employees
select count(*) from Employee

-- Group by with multiple aggregates
select department, count(*), avg(salary), max(salary) from Employee group-by department

-- Group by with having clause
select department, count(*) from Employee group-by department having count(*) > 5
```

## API Reference

### Creating Queries

```go
// Parse and interpret in one step
query, err := interpreter.NewQuery(sqlString, resources)

// Parse only (syntax validation)
parsed, err := parser.NewQuery(sqlString, logger)

// Interpret a pre-parsed query
query, err := interpreter.NewFromQuery(parsedQuery.Query(), resources)
```

### Query Methods

```go
// Filtering
query.Match(obj interface{}) bool                                    // Test single object
query.Filter(list []interface{}, onlySelectedColumns bool) []interface{} // Filter list

// Aggregation
query.IsAggregate() bool                                             // Check if aggregate query
query.Aggregate(list []interface{}) []map[string]interface{}         // Execute aggregation

// Properties
query.Properties() []ifs.IProperty          // Selected columns as properties
query.PropertiesMap() map[string]ifs.IProperty
query.RootType() *l8reflect.L8Node          // Resolved root type
query.Criteria() ifs.IExpression            // WHERE clause expression tree

// Sorting
query.SortBy() string                       // Sort-by column name
query.SortByValue(v interface{}) interface{} // Extract sort key from object
query.SortByProperty() *properties.Property
query.Descending() bool

// Pagination
query.Limit() int32
query.Page() int32

// Flags
query.MatchCase() bool
query.MapReduce() bool

// Utility
query.Hash() string                         // MD5 hash for caching
query.String() string                       // Reconstructed query string
query.Text() string                         // Original query text
query.ValueForParameter(name string) string // Extract value from WHERE clause
query.KeyOf() string                        // Extract key value
```

## Testing

### Running Tests
```bash
cd go

# Run all tests
go test ./... -v

# Run with coverage report
go test -v -coverpkg=./gsql/... -coverprofile=cover.html ./... --failfast

# Full test suite (fetches deps, runs tests, opens coverage)
./test.sh
```

### Test Categories
- **Parser Tests** (`Parser_test.go`) - Query parsing and syntax validation
- **Interpreter Tests** (`Interpreter_test.go`) - Query execution, matching, and filtering
- **Aggregate Tests** (`Aggregate_test.go`) - Aggregate functions, group-by, having
- **Special Case Tests** (`TestSpecialCase_test.go`) - Edge cases and boundary conditions

## Project Structure

```
l8ql/
├── go/
│   ├── gsql/
│   │   ├── parser/                 # Stage 1: Query string -> L8Query protobuf
│   │   │   ├── Query.go           # Main parser entry point
│   │   │   ├── Aggregate.go       # Aggregate function parsing (count, sum, avg, min, max)
│   │   │   ├── Expression.go      # WHERE clause expression parsing
│   │   │   ├── Condition.go       # AND/OR condition chain parsing
│   │   │   └── Comparator.go      # Comparison operator parsing
│   │   │
│   │   └── interpreter/            # Stage 2: L8Query protobuf -> executable query
│   │       ├── Query.go           # Main interpreter, Match/Filter/Aggregate
│   │       ├── Expression.go      # Interpreted expression tree
│   │       ├── Condition.go       # Interpreted condition chain
│   │       ├── Comparator.go      # Interpreted comparison with property resolution
│   │       ├── Accumulator.go     # Aggregate state tracking (count/sum/avg/min/max)
│   │       └── comparators/       # Type-aware comparison operators
│   │           ├── Equal.go       # Equality with wildcard and bool support
│   │           ├── NotEqual.go
│   │           ├── GreaterThan.go
│   │           ├── GreaterThanOrEqual.go
│   │           ├── LessThan.go
│   │           ├── LessThanOrEqual.go
│   │           ├── In.go          # Membership test
│   │           └── NotIn.go       # Negative membership test
│   │
│   ├── tests/                      # All tests
│   │   ├── TestInit.go            # Test infrastructure setup
│   │   ├── Parser_test.go
│   │   ├── Interpreter_test.go
│   │   ├── Aggregate_test.go
│   │   └── TestSpecialCase_test.go
│   │
│   ├── test.sh                     # Test runner with coverage
│   ├── go.mod
│   └── vendor/                     # Vendored dependencies
│
├── LICENSE                         # GPL v3
└── README.md
```

## Dependencies

### Direct
| Dependency | Purpose |
|-----------|---------|
| [l8types](https://github.com/saichler/l8types) | Core interfaces (`IQuery`, `IExpression`, `IResources`) and L8Query protobuf definition |
| [l8reflect](https://github.com/saichler/l8reflect) | Reflection utilities, introspection, and property access |
| [l8test](https://github.com/saichler/l8test) | Test infrastructure and test models |
| [l8pollaris](https://github.com/saichler/l8pollaris) | Plugin loading support |

### Indirect
| Dependency | Purpose |
|-----------|---------|
| google.golang.org/protobuf | Protobuf runtime |
| l8bus, l8services, l8srlz, l8utils | Transitive via l8test |

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Add tests for new functionality
5. Run the test suite (`cd go && go test ./... -v`)
6. Commit your changes
7. Push and open a Pull Request

## License

This project is licensed under the GNU General Public License v3.0 - see the [LICENSE](LICENSE) file for details.

## Support

- **Issues**: [GitHub Issues](https://github.com/saichler/l8ql/issues)

---

**L8QL** - Simplifying graph data querying with familiar SQL syntax.
