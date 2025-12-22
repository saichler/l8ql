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

// NotEqual implements the not-equal (!=) comparison operator.
// It supports string, signed integer, and unsigned integer types.
type NotEqual struct {
	compares map[reflect.Kind]func(interface{}, interface{}) bool
}

// NewNotEqual creates a new NotEqual comparator with type-specific matcher functions.
func NewNotEqual() *NotEqual {
	c := &NotEqual{}
	c.compares = make(map[reflect.Kind]func(interface{}, interface{}) bool)
	c.compares[reflect.String] = noteqStringMatcher
	c.compares[reflect.Int] = noteqIntMatcher
	c.compares[reflect.Int8] = noteqIntMatcher
	c.compares[reflect.Int16] = noteqIntMatcher
	c.compares[reflect.Int32] = noteqIntMatcher
	c.compares[reflect.Int64] = noteqIntMatcher
	c.compares[reflect.Uint] = noteqUintMatcher
	c.compares[reflect.Uint8] = noteqUintMatcher
	c.compares[reflect.Uint16] = noteqUintMatcher
	c.compares[reflect.Uint32] = noteqUintMatcher
	c.compares[reflect.Uint64] = noteqUintMatcher
	return c
}

// Compare evaluates inequality between left and right values.
func (notequal *NotEqual) Compare(left, right interface{}) bool {
	return Compare(left, right, notequal.compares, "Not Equal")
}

// noteqStringMatcher compares two string values for inequality.
func noteqStringMatcher(left, right interface{}) bool {
	aside := removeSingleQuote(strings.ToLower(left.(string)))
	zside := removeSingleQuote(strings.ToLower(right.(string)))
	return aside != zside
}

// noteqIntMatcher compares signed integer values for inequality.
func noteqIntMatcher(left, right interface{}) bool {
	aside, ok := getInt64(left)
	if !ok {
		return false
	}
	zside, ok := getInt64(right)
	if !ok {
		return false
	}
	return aside != zside
}

// noteqUintMatcher compares unsigned integer values for inequality.
func noteqUintMatcher(left, right interface{}) bool {
	aside, ok := getUint64(left)
	if !ok {
		return false
	}
	zside, ok := getUint64(right)
	if !ok {
		return false
	}
	return aside != zside
}
