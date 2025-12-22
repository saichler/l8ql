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
package comparators

import (
	"reflect"
	"strings"
)

// LessThanOrEqual implements the less-than-or-equal (<=) comparison operator.
// It supports string, signed integer, and unsigned integer types.
type LessThanOrEqual struct {
	compares map[reflect.Kind]func(interface{}, interface{}) bool
}

// NewLessThanOrEqual creates a new LessThanOrEqual comparator with type-specific matcher functions.
func NewLessThanOrEqual() *LessThanOrEqual {
	c := &LessThanOrEqual{}
	c.compares = make(map[reflect.Kind]func(interface{}, interface{}) bool)
	c.compares[reflect.String] = lteqStringMatcher
	c.compares[reflect.Int] = lteqIntMatcher
	c.compares[reflect.Int8] = lteqIntMatcher
	c.compares[reflect.Int16] = lteqIntMatcher
	c.compares[reflect.Int32] = lteqIntMatcher
	c.compares[reflect.Int64] = lteqIntMatcher
	c.compares[reflect.Uint] = lteqUintMatcher
	c.compares[reflect.Uint8] = lteqUintMatcher
	c.compares[reflect.Uint16] = lteqUintMatcher
	c.compares[reflect.Uint32] = lteqUintMatcher
	c.compares[reflect.Uint64] = lteqUintMatcher
	return c
}

// Compare evaluates whether left is less than or equal to right.
func (lteq *LessThanOrEqual) Compare(left, right interface{}) bool {
	return Compare(left, right, lteq.compares, "Less Than Or Equal")
}

// lteqStringMatcher compares two string values lexicographically.
func lteqStringMatcher(left, right interface{}) bool {
	aside := removeSingleQuote(strings.ToLower(left.(string)))
	zside := removeSingleQuote(strings.ToLower(right.(string)))
	return aside <= zside
}

// lteqIntMatcher compares signed integer values.
func lteqIntMatcher(left, right interface{}) bool {
	aside, ok := getInt64(left)
	if !ok {
		return false
	}
	zside, ok := getInt64(right)
	if !ok {
		return false
	}
	return aside <= zside
}

// lteqUintMatcher compares unsigned integer values.
func lteqUintMatcher(left, right interface{}) bool {
	aside, ok := getUint64(left)
	if !ok {
		return false
	}
	zside, ok := getUint64(right)
	if !ok {
		return false
	}
	return aside <= zside
}
