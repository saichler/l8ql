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

// LessThan implements the less-than (<) comparison operator.
// It supports string, signed integer, and unsigned integer types.
type LessThan struct {
	compares map[reflect.Kind]func(interface{}, interface{}) bool
}

// NewLessThan creates a new LessThan comparator with type-specific matcher functions.
func NewLessThan() *LessThan {
	c := &LessThan{}
	c.compares = make(map[reflect.Kind]func(interface{}, interface{}) bool)
	c.compares[reflect.String] = ltStringMatcher
	c.compares[reflect.Int] = ltIntMatcher
	c.compares[reflect.Int8] = ltIntMatcher
	c.compares[reflect.Int16] = ltIntMatcher
	c.compares[reflect.Int32] = ltIntMatcher
	c.compares[reflect.Int64] = ltIntMatcher
	c.compares[reflect.Uint] = ltUintMatcher
	c.compares[reflect.Uint8] = ltUintMatcher
	c.compares[reflect.Uint16] = ltUintMatcher
	c.compares[reflect.Uint32] = ltUintMatcher
	c.compares[reflect.Uint64] = ltUintMatcher
	return c
}

// Compare evaluates whether left is less than right.
func (lt *LessThan) Compare(left, right interface{}) bool {
	return Compare(left, right, lt.compares, "Less Than")
}

// ltStringMatcher compares two string values lexicographically.
func ltStringMatcher(left, right interface{}) bool {
	aside := removeSingleQuote(strings.ToLower(left.(string)))
	zside := removeSingleQuote(strings.ToLower(right.(string)))
	return aside < zside
}

// ltIntMatcher compares signed integer values.
func ltIntMatcher(left, right interface{}) bool {
	aside, ok := getInt64(left)
	if !ok {
		return false
	}
	zside, ok := getInt64(right)
	if !ok {
		return false
	}
	return aside < zside
}

// ltUintMatcher compares unsigned integer values.
func ltUintMatcher(left, right interface{}) bool {
	aside, ok := getUint64(left)
	if !ok {
		return false
	}
	zside, ok := getUint64(right)
	if !ok {
		return false
	}
	return aside < zside
}
