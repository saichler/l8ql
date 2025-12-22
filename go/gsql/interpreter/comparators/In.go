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

// IN implements the membership (in) comparison operator.
// It checks if the left value exists within a list of values on the right.
// The list is specified in bracket notation: [val1,val2,val3]
type IN struct {
	compares map[reflect.Kind]func(interface{}, interface{}) bool
}

// NewIN creates a new IN comparator with type-specific matcher functions.
func NewIN() *IN {
	c := &IN{}
	c.compares = make(map[reflect.Kind]func(interface{}, interface{}) bool)
	c.compares[reflect.String] = inStringMatcher
	c.compares[reflect.Int] = inIntMatcher
	c.compares[reflect.Int8] = inIntMatcher
	c.compares[reflect.Int16] = inIntMatcher
	c.compares[reflect.Int32] = inIntMatcher
	c.compares[reflect.Int64] = inIntMatcher
	c.compares[reflect.Uint] = inUintMatcher
	c.compares[reflect.Uint8] = inUintMatcher
	c.compares[reflect.Uint16] = inUintMatcher
	c.compares[reflect.Uint32] = inUintMatcher
	c.compares[reflect.Uint64] = inUintMatcher
	return c
}

// Compare evaluates whether left is in the list specified by right.
func (in *IN) Compare(left, right interface{}) bool {
	return Compare(left, right, in.compares, "In")
}

// inStringMatcher checks if a string value is in a list of strings.
func inStringMatcher(left, right interface{}) bool {
	aside := removeSingleQuote(strings.ToLower(left.(string)))
	zsideList := strings.ToLower(right.(string))
	values := getInStringList(zsideList)
	for _, v := range values {
		if aside == v {
			return true
		}
	}
	return false
}

// inIntMatcher checks if a signed integer value is in a list of integers.
func inIntMatcher(left, right interface{}) bool {
	aside, ok := getInt64(left)
	if !ok {
		return false
	}

	zsideList := strings.ToLower(right.(string))

	values := getInStringList(zsideList)
	for _, v := range values {
		intV, e := strconv.Atoi(v)
		if e != nil {
			return false
		}
		if aside == int64(intV) {
			return true
		}
	}
	return false
}

// inUintMatcher checks if an unsigned integer value is in a list of integers.
func inUintMatcher(left, right interface{}) bool {
	aside, ok := getUint64(left)
	if !ok {
		return false
	}

	zsideList := strings.ToLower(right.(string))

	values := getInStringList(zsideList)
	for _, v := range values {
		intV, e := strconv.Atoi(v)
		if e != nil {
			return false
		}
		if aside == uint64(intV) {
			return true
		}
	}
	return false
}

// getInStringList extracts the list of values from a bracket-enclosed string.
// E.g., "[a,b,c]" returns ["a", "b", "c"]
func getInStringList(str string) []string {
	index := strings.Index(str, "[")
	index2 := strings.Index(str, "]")
	lst := str[index+1 : index2]
	values := strings.Split(lst, ",")
	result := make([]string, 0)
	for _, v := range values {
		result = append(result, removeSingleQuote(v))
	}
	return result
}
