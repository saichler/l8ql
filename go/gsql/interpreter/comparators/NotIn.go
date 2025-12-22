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
	"strconv"
	"strings"
)

// NotIN implements the negative membership (not in) comparison operator.
// It checks if the left value does NOT exist within a list of values on the right.
// The list is specified in bracket notation: [val1,val2,val3]
type NotIN struct {
	compares map[reflect.Kind]func(interface{}, interface{}) bool
}

// NewNotIN creates a new NotIN comparator with type-specific matcher functions.
func NewNotIN() *NotIN {
	c := &NotIN{}
	c.compares = make(map[reflect.Kind]func(interface{}, interface{}) bool)
	c.compares[reflect.String] = notinStringMatcher
	c.compares[reflect.Int] = notinIntMatcher
	c.compares[reflect.Int8] = notinIntMatcher
	c.compares[reflect.Int16] = notinIntMatcher
	c.compares[reflect.Int32] = notinIntMatcher
	c.compares[reflect.Int64] = notinIntMatcher
	c.compares[reflect.Uint] = notinUintMatcher
	c.compares[reflect.Uint8] = notinUintMatcher
	c.compares[reflect.Uint16] = notinUintMatcher
	c.compares[reflect.Uint32] = notinUintMatcher
	c.compares[reflect.Uint64] = notinUintMatcher
	return c
}

// Compare evaluates whether left is NOT in the list specified by right.
func (in *NotIN) Compare(left, right interface{}) bool {
	return Compare(left, right, in.compares, "In")
}

// notinStringMatcher checks if a string value is NOT in a list of strings.
func notinStringMatcher(left, right interface{}) bool {
	aside := removeSingleQuote(strings.ToLower(left.(string)))
	zsideList := strings.ToLower(right.(string))
	values := getInStringList(zsideList)
	for _, v := range values {
		if aside == v {
			return false
		}
	}
	return true
}

// notinIntMatcher checks if a signed integer value is NOT in a list of integers.
func notinIntMatcher(left, right interface{}) bool {
	aside, ok := getInt64(left)
	if !ok {
		return true
	}

	zsideList := strings.ToLower(right.(string))

	values := getInStringList(zsideList)
	for _, v := range values {
		intV, e := strconv.Atoi(v)
		if e != nil {
			return true
		}
		if aside == int64(intV) {
			return false
		}
	}
	return true
}

// notinUintMatcher checks if an unsigned integer value is NOT in a list of integers.
func notinUintMatcher(left, right interface{}) bool {
	aside, ok := getUint64(left)
	if !ok {
		return true
	}

	zsideList := strings.ToLower(right.(string))

	values := getInStringList(zsideList)
	for _, v := range values {
		intV, e := strconv.Atoi(v)
		if e != nil {
			return true
		}
		if aside == uint64(intV) {
			return false
		}
	}
	return true
}
