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

// GreaterThanOrEqual implements the greater-than-or-equal (>=) comparison operator.
// It supports string, signed integer, and unsigned integer types.
type GreaterThanOrEqual struct {
	compares map[reflect.Kind]func(interface{}, interface{}) bool
}

// NewGreaterThanOrEqual creates a new GreaterThanOrEqual comparator with type-specific matcher functions.
func NewGreaterThanOrEqual() *GreaterThanOrEqual {
	c := &GreaterThanOrEqual{}
	c.compares = make(map[reflect.Kind]func(interface{}, interface{}) bool)
	c.compares[reflect.String] = gteqStringMatcher
	c.compares[reflect.Int] = gteqIntMatcher
	c.compares[reflect.Int8] = gteqIntMatcher
	c.compares[reflect.Int16] = gteqIntMatcher
	c.compares[reflect.Int32] = gteqIntMatcher
	c.compares[reflect.Int64] = gteqIntMatcher
	c.compares[reflect.Uint] = gteqUintMatcher
	c.compares[reflect.Uint8] = gteqUintMatcher
	c.compares[reflect.Uint16] = gteqUintMatcher
	c.compares[reflect.Uint32] = gteqUintMatcher
	c.compares[reflect.Uint64] = gteqUintMatcher
	return c
}

// Compare evaluates whether left is greater than or equal to right.
func (gteq *GreaterThanOrEqual) Compare(left, right interface{}) bool {
	return Compare(left, right, gteq.compares, "Greater Than Or Equal")
}

// gteqStringMatcher compares two string values lexicographically.
func gteqStringMatcher(left, right interface{}) bool {
	aside := removeSingleQuote(strings.ToLower(left.(string)))
	zside := removeSingleQuote(strings.ToLower(right.(string)))
	return aside >= zside
}

// gteqIntMatcher compares signed integer values.
func gteqIntMatcher(left, right interface{}) bool {
	aside, ok := getInt64(left)
	if !ok {
		return false
	}
	zside, ok := getInt64(right)
	if !ok {
		return false
	}
	return aside >= zside
}

// gteqUintMatcher compares unsigned integer values.
func gteqUintMatcher(left, right interface{}) bool {
	aside, ok := getUint64(left)
	if !ok {
		return false
	}
	zside, ok := getUint64(right)
	if !ok {
		return false
	}
	return aside >= zside
}
