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

// Accumulator.go tracks running state for aggregate function computation.
// Supports count, sum, avg, min, and max over numeric types (int32, int64, float32, float64).
package interpreter

import (
	"math"
)

// Accumulator tracks running state for a single aggregate function.
type Accumulator struct {
	fn    string  // "count", "sum", "avg", "min", "max"
	count int64   // Number of values added
	sum   float64 // Running sum (for sum, avg)
	min   float64 // Running minimum
	max   float64 // Running maximum
	hasValue bool // Whether any non-nil value has been added
}

// NewAccumulator creates a new Accumulator for the given function name.
func NewAccumulator(fn string) *Accumulator {
	return &Accumulator{
		fn:  fn,
		min: math.MaxFloat64,
		max: -math.MaxFloat64,
	}
}

// Add incorporates a value into the accumulator.
// For count(*), pass nil to count all records.
// Handles int32, int64, float32, float64 value types.
func (a *Accumulator) Add(value interface{}) {
	a.count++

	if value == nil {
		return
	}

	num, ok := toFloat64(value)
	if !ok {
		return
	}

	a.hasValue = true
	a.sum += num
	if num < a.min {
		a.min = num
	}
	if num > a.max {
		a.max = num
	}
}

// Result returns the final computed value for this accumulator.
func (a *Accumulator) Result() interface{} {
	switch a.fn {
	case "count":
		return a.count
	case "sum":
		return a.sum
	case "avg":
		if a.count == 0 {
			return float64(0)
		}
		return a.sum / float64(a.count)
	case "min":
		if !a.hasValue {
			return float64(0)
		}
		return a.min
	case "max":
		if !a.hasValue {
			return float64(0)
		}
		return a.max
	}
	return nil
}

// toFloat64 converts a numeric value to float64 for accumulation.
func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case int:
		return float64(v), true
	}
	return 0, false
}
